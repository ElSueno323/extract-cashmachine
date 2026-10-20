#!/usr/bin/env python3
"""Export des transactions ShopMaster vers Excel.

Demande les sites et la période, télécharge le CSV de chaque caisse
(séparateur « ; »), puis écrit un fichier .xlsx par site.

Usage :
    python export_transactions.py
    TransactionsCaisses.exe
"""

from __future__ import annotations

import csv
import io
import sys
from datetime import datetime
from pathlib import Path
from urllib.error import HTTPError, URLError
from urllib.parse import quote
from urllib.request import Request, urlopen

try:
    from openpyxl import Workbook
    from openpyxl.styles import Font
    from openpyxl.utils import get_column_letter
except ImportError:
    print("Le module openpyxl est manquant.")
    print("Installez-le avec : python -m pip install -r requirements.txt")
    sys.exit(1)

SITES: list[dict] = []
TIMEOUT_SECONDS = 20
CSV_PATH = ""


def application_dir() -> Path:
    """Dossier du script, ou de l'exécutable une fois empaqueté."""
    if getattr(sys, "frozen", False):
        return Path(sys.executable).resolve().parent
    return Path(__file__).resolve().parent


OUTPUT_DIR = application_dir()


def load_settings() -> None:
    """Lit les caisses et les paramètres depuis le fichier .env."""
    global SITES, TIMEOUT_SECONDS, CSV_PATH

    candidates = [Path.cwd() / ".env", application_dir() / ".env"]
    env_path = next((path for path in candidates if path.is_file()), None)
    if env_path is None:
        print("Fichier .env introuvable.")
        raise SystemExit(1)

    values: dict[str, str] = {}
    for raw in env_path.read_text(encoding="utf-8").splitlines():
        line = raw.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        key, value = line.split("=", 1)
        values[key.strip()] = value.strip().strip('"').strip("'")

    sites = []
    index = 1
    while True:
        raw_site = values.get(f"SITE_{index}", "").strip()
        if not raw_site:
            break
        parts = [part.strip() for part in raw_site.split("|", 2)]
        if len(parts) != 3 or not all(parts):
            print(f"SITE_{index} est invalide dans .env.")
            raise SystemExit(1)
        sites.append(
            {"key": str(index), "slug": parts[0], "name": parts[1], "ip": parts[2]}
        )
        index += 1

    if not sites:
        print("Aucune caisse dans .env (SITE_1, SITE_2, …).")
        raise SystemExit(1)

    csv_path = values.get("CSV_PATH", "").strip()
    if not csv_path.startswith("/"):
        print("CSV_PATH manquant ou invalide dans .env.")
        raise SystemExit(1)

    timeout = values.get("HTTP_TIMEOUT_SECONDS", "").strip()
    if not timeout.isdigit() or int(timeout) <= 0:
        print("HTTP_TIMEOUT_SECONDS invalide dans .env.")
        raise SystemExit(1)

    SITES = sites
    CSV_PATH = csv_path
    TIMEOUT_SECONDS = int(timeout)


def prompt(text: str) -> str:
    try:
        return input(text).strip()
    except (EOFError, KeyboardInterrupt):
        print("\nInterrompu.")
        raise SystemExit(1) from None


def choose_sites() -> list[dict]:
    print("Sites disponibles :")
    for site in SITES:
        print(f"  {site['key']}. {site['name']} ({site['ip']})")
    print("  4. Tous les sites")
    print()
    print("Un ou plusieurs numéros, séparés par une virgule ou un espace.")
    print("Exemples : 1    1,3    2 3    4")

    by_name = {}
    for site in SITES:
        by_name[site["key"]] = site
        by_name[site["slug"]] = site
        by_name[site["name"].lower()] = site

    while True:
        raw = prompt("Site(s) : ")
        tokens = [
            token
            for token in raw.replace(";", ",").replace(" ", ",").split(",")
            if token.strip()
        ]
        tokens = [token.strip() for token in tokens]
        if not tokens:
            print("Choix vide. Réessayez.")
            continue

        lowered = {token.lower() for token in tokens}
        if "4" in tokens or "tous" in lowered or "all" in lowered:
            return list(SITES)

        selected = []
        unknown = []
        seen = set()
        for token in tokens:
            site = by_name.get(token.lower()) or by_name.get(token)
            if site is None:
                unknown.append(token)
                continue
            if site["ip"] not in seen:
                selected.append(site)
                seen.add(site["ip"])

        if unknown:
            print(f"Choix inconnu : {', '.join(unknown)}. Réessayez.")
            continue
        if not selected:
            print("Aucun site sélectionné. Réessayez.")
            continue
        return selected


def parse_date(label: str) -> datetime:
    while True:
        raw = prompt(f"{label} (JJ/MM/AAAA) : ")
        try:
            return datetime.strptime(raw, "%d/%m/%Y")
        except ValueError:
            print("Date invalide. Exemple : 11/09/2026.")


def choose_dates() -> tuple[datetime, datetime]:
    while True:
        start = parse_date("Date de début")
        end = parse_date("Date de fin")
        if start <= end:
            return start, end
        print("La date de début doit être antérieure ou égale à la date de fin.")


def build_url(ip: str, start: datetime, end: datetime) -> str:
    start_q = quote(start.strftime("%d/%m/%Y"), safe="")
    end_q = quote(end.strftime("%d/%m/%Y"), safe="")
    return f"http://{ip}{CSV_PATH}?startdate={start_q}&enddate={end_q}"


def download_csv(url: str) -> bytes:
    request = Request(url, headers={"User-Agent": "cashmachine-transactions/1.0"})
    with urlopen(request, timeout=TIMEOUT_SECONDS) as response:
        return response.read()


def decode_csv(data: bytes) -> str:
    if not data or not data.strip():
        return ""
    for encoding in ("utf-8-sig", "cp1252", "latin-1"):
        try:
            return data.decode(encoding)
        except UnicodeDecodeError:
            continue
    return data.decode("latin-1", errors="replace")


def rows_from_csv(text: str) -> list[list[str]]:
    """Découpe chaque ligne sur le séparateur « ; »."""
    if not text.strip():
        return []
    reader = csv.reader(io.StringIO(text), delimiter=";")
    return [row for row in reader if any(cell.strip() for cell in row)]


def write_xlsx(path: Path, rows: list[list[str]]) -> None:
    workbook = Workbook()
    sheet = workbook.active
    sheet.title = "Transactions"
    for row in rows:
        sheet.append(row)

    if rows:
        for cell in sheet[1]:
            cell.font = Font(bold=True)
        sheet.auto_filter.ref = sheet.dimensions
        sheet.freeze_panes = "A2"
        sample = rows[:200]
        width_by_col = [10] * max(len(row) for row in sample)
        for row in sample:
            for index, value in enumerate(row):
                width_by_col[index] = max(
                    width_by_col[index], min(len(value) + 2, 60)
                )
        for index, width in enumerate(width_by_col, start=1):
            sheet.column_dimensions[get_column_letter(index)].width = width

    path.parent.mkdir(parents=True, exist_ok=True)
    workbook.save(path)


def skip_site(site: dict, reason: str) -> None:
    print(f"  {site['name']} ignoré : {reason}")


def export_site(site: dict, start: datetime, end: datetime) -> Path | None:
    url = build_url(site["ip"], start, end)
    print(f"\n{site['name']} — téléchargement…")
    print(f"  {url}")
    try:
        data = download_csv(url)
        text = decode_csv(data)
        if text.lstrip().startswith("<"):
            skip_site(site, "la caisse n'a pas renvoyé un CSV.")
            return None

        rows = rows_from_csv(text)
        stamp = f"{start.strftime('%d-%m-%Y')}_{end.strftime('%d-%m-%Y')}"
        path = OUTPUT_DIR / f"{site['slug']}_{stamp}.xlsx"
        write_xlsx(path, rows)
    except HTTPError as exc:
        skip_site(site, f"erreur HTTP {exc.code}.")
        return None
    except URLError as exc:
        skip_site(site, f"impossible de joindre {site['ip']} ({exc.reason}).")
        return None
    except TimeoutError:
        skip_site(site, "délai dépassé.")
        return None
    except OSError as exc:
        skip_site(site, f"accès impossible ({exc}).")
        return None
    except Exception as exc:
        skip_site(site, str(exc))
        return None

    transactions = max(len(rows) - 1, 0) if rows else 0
    if not rows:
        print("  Aucune transaction sur cette période.")
    else:
        print(f"  {transactions} transaction(s), {len(rows)} ligne(s) au total.")
    print(f"  Fichier : {path}")
    return path


def configure_console() -> None:
    if sys.platform == "win32":
        try:
            import ctypes

            ctypes.windll.kernel32.SetConsoleOutputCP(65001)
            ctypes.windll.kernel32.SetConsoleCP(65001)
        except OSError:
            pass
    for stream in (sys.stdout, sys.stdin, sys.stderr):
        if hasattr(stream, "reconfigure"):
            try:
                stream.reconfigure(encoding="utf-8")
            except (OSError, ValueError):
                pass


def pause_before_exit() -> None:
    if not getattr(sys, "frozen", False):
        return
    try:
        input("\nAppuyez sur Entrée pour fermer...")
    except (EOFError, KeyboardInterrupt):
        pass


def main() -> None:
    configure_console()
    print("Export des transactions ShopMaster")
    print("==================================")
    while True:
        sites = choose_sites()
        start, end = choose_dates()
        names = ", ".join(site["name"] for site in sites)
        print(f"\nExtraction : {names}")
        print(
            f"Période    : {start.strftime('%d/%m/%Y')} → {end.strftime('%d/%m/%Y')}"
        )

        ok = 0
        for index, site in enumerate(sites):
            try:
                saved = export_site(site, start, end)
            except Exception as exc:
                skip_site(site, str(exc))
                saved = None
            if saved:
                ok += 1
            elif index < len(sites) - 1:
                print("  Passage à la caisse suivante.")

        print(f"\nTerminé : {ok}/{len(sites)} fichier(s) dans {OUTPUT_DIR}")
        again = prompt("\nNouvelle extraction ? (o/n) : ").lower()
        if again not in {"o", "oui", "y", "yes"}:
            break
    print("Au revoir.")


if __name__ == "__main__":
    try:
        load_settings()
        main()
    except SystemExit:
        raise
    except Exception as exc:
        print(f"\nErreur : {exc}")
    finally:
        pause_before_exit()
