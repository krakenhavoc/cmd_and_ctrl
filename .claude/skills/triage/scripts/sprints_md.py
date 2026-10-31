#!/usr/bin/env -S python3 -I
"""Write a sprint the way docs/sprints.md and the tracker issues record one.

  sprints_md.py render <spec.json> [--part tracker|section|row]
  sprints_md.py add-sprint <sprints.md> <spec.json> [--dry-run]
  sprints_md.py mark-done <sprints.md> <SNN> <YYYY-MM-DD> "<what closed it>" [--dry-run]
  sprints_md.py apply <file> <edits.json> [--dry-run]

render prints the tracker issue body, the sprints.md section and the index row
(or one of them). add-sprint puts the row at the end of the sprint index and the
section before "## Backlog". mark-done sets a sprint's index row to **done** and
puts "**Done** (<date> triage). <note>" ahead of its old status, which stays
as "Before that: …". apply makes exact-text edits, for status notes and the
Backlog prose.

Spec (JSON); see references/sprints-doc.md for a full example:
  key            "S81"
  name           theme words; the milestone is "S81 — <name>"
  phase          7
  tracker        the tracker issue number (section and row need it; the tracker body doesn't)
  goal           one or two sentences
  date           the triage that opened it, YYYY-MM-DD
  opened         optional; replaces "Opened at the <date> triage."
  context        optional; follows the opened sentence
  groups         [{"heading": optional line, "items": [{"n": 4326, "text": "...", "done": false}]}]
                 an item without "n" is a plain bullet
  exit_criteria  ["...", ...]
  status         optional; default "**Planned** (<date> triage)."
  row_status     optional; default "planned"

apply's edits file is a JSON list of {"old": ..., "new": ...}, {"after": ..., "insert": ...}
or {"before": ..., "insert": ...}. Each anchor must occur exactly once. An edit
that is already in the file is skipped, so every command here is safe to rerun.
--dry-run prints a diff and writes nothing.
"""

import argparse
import difflib
import json
import os
import re
import subprocess
import sys
from pathlib import Path

HERE = Path(__file__).resolve().parent
ROOT = subprocess.run(["git", "-C", str(HERE), "rev-parse", "--show-toplevel"],
                      capture_output=True, text=True, check=True).stdout.strip()


def issue_url_base() -> str:
    repo = os.environ.get("TRIAGE_REPO")
    if not repo:
        r = subprocess.run(["gh", "repo", "view", "--json", "nameWithOwner", "--jq", ".nameWithOwner"],
                           cwd=ROOT, capture_output=True, text=True)
        repo = r.stdout.strip()
    if not repo:
        remote = subprocess.run(["git", "-C", ROOT, "remote", "get-url", "origin"],
                                capture_output=True, text=True).stdout.strip()
        m = re.search(r"github\.com[:/](.+?)(?:\.git)?$", remote)
        if not m:
            sys.exit("cannot tell the repository: set TRIAGE_REPO=owner/name")
        repo = m.group(1)
    return f"https://github.com/{repo}/issues/"


def link(n: int, base: str) -> str:
    return f"[#{n}]({base}{n})"


def sentence(s: str) -> str:
    s = s.strip()
    return s if s.endswith((".", "!", "?", ")")) else s + "."


def checklist(spec, base, style: str) -> str:
    blocks = []
    for g in spec.get("groups", []):
        lines = []
        for it in g["items"]:
            if "n" not in it:
                lines.append(f"- {it['text']}")
                continue
            box = "x" if it.get("done") else " "
            ref = link(it["n"], base) + ":" if style == "doc" else f"#{it['n']}"
            lines.append(f"- [{box}] {ref} {it['text']}")
        body = "\n".join(lines)
        blocks.append(f"{g['heading']}\n\n{body}" if g.get("heading") else body)
    return "\n\n".join(blocks)


def opened(spec) -> str:
    text = spec.get("opened") or f"Opened at the {spec['date']} triage."
    return f"{sentence(text)} {spec['context'].strip()}" if spec.get("context") else sentence(text)


def render_tracker(spec, base) -> str:
    goal = sentence(spec["goal"])
    parts = [f"Tracker for {spec['key']}, opened at the {spec['date']} triage. **Goal:** {goal}",
             checklist(spec, base, "tracker")]
    if spec.get("context"):
        parts.append(spec["context"].strip())
    crit = "\n".join(f"{i}. {c}" for i, c in enumerate(spec["exit_criteria"], 1))
    parts.append(f"## Exit criteria\n{crit}")
    return "\n\n".join(parts) + "\n"


def need_tracker(spec) -> int:
    t = spec.get("tracker")
    if not t:
        sys.exit("the spec has no tracker number yet: create the tracker first (new-sprint.sh)")
    return int(t)


def render_section(spec, base) -> str:
    t = need_tracker(spec)
    crit = "\n".join(f"{i}. {c}" for i, c in enumerate(spec["exit_criteria"], 1))
    status = spec.get("status") or f"**Planned** ({spec['date']} triage)."
    return (f"## {spec['key']} — {spec['name']}\n\n"
            f"**Phase:** {spec['phase']} · **Goal:** {sentence(spec['goal'])} Tracking issue {link(t, base)}.\n\n"
            f"{opened(spec)}\n\n"
            f"{checklist(spec, base, 'doc')}\n\n"
            f"### Exit criteria\n\n{crit}\n\n"
            f"### Status\n\n{status}\n\n---\n\n")


def render_row(spec, base) -> str:
    t = need_tracker(spec)
    return (f"| {spec['key']:<7} | {spec['name']:<68} | {str(spec['phase']):<5} | {link(t, base)} | "
            f"{'—':<10} | {spec.get('row_status', 'planned'):<11} |")


def index_bounds(text: str):
    m = re.search(r"^## Sprint index\n", text, re.M)
    if not m:
        sys.exit("no '## Sprint index' heading")
    end = re.search(r"^## ", text[m.end():], re.M)
    return m.end(), m.end() + end.start() if end else len(text)


def show_or_write(path: Path, before: str, after: str, dry: bool):
    if before == after:
        print(f"{path}: nothing to change")
        return
    if dry:
        sys.stdout.writelines(difflib.unified_diff(before.splitlines(True), after.splitlines(True),
                                                   str(path), str(path) + " (new)", n=1))
        return
    path.write_text(after, encoding="utf-8")
    print(f"{path}: written")


def cmd_render(a):
    spec = json.loads(Path(a.spec).read_text(encoding="utf-8"))
    base = issue_url_base()
    parts = {"tracker": lambda: render_tracker(spec, base),
             "section": lambda: render_section(spec, base),
             "row": lambda: render_row(spec, base) + "\n"}
    if a.part:
        sys.stdout.write(parts[a.part]())
        return
    for name in ("tracker", "section", "row"):
        print(f"===== {name}")
        sys.stdout.write(parts[name]())


def cmd_add_sprint(a):
    path = Path(a.file)
    spec = json.loads(Path(a.spec).read_text(encoding="utf-8"))
    base = issue_url_base()
    text = before = path.read_text(encoding="utf-8")
    key = spec["key"]

    lo, hi = index_bounds(text)
    if re.search(rf"^\| {re.escape(key)}\s+\|", text[lo:hi], re.M):
        print(f"index row for {key} is already there")
    else:
        rows = list(re.finditer(r"^\| S\d+(?:\.\d+)?\s+\|.*$", text[lo:hi], re.M))
        if not rows:
            sys.exit("no sprint rows in the index")
        at = lo + rows[-1].end()
        text = text[:at] + "\n" + render_row(spec, base) + text[at:]

    if re.search(rf"^## {re.escape(key)} — ", text, re.M):
        print(f"section for {key} is already there")
    else:
        m = re.search(r"^## Backlog", text, re.M)
        if not m:
            sys.exit("no '## Backlog' heading to put the section before; place it with `apply`")
        text = text[:m.start()] + render_section(spec, base) + text[m.start():]
    show_or_write(path, before, text, a.dry_run)


def cmd_mark_done(a):
    path = Path(a.file)
    text = before = path.read_text(encoding="utf-8")
    key = a.key

    lo, hi = index_bounds(text)
    m = re.search(rf"^\| {re.escape(key)}\s+\|.*$", text[lo:hi], re.M)
    if not m:
        sys.exit(f"no index row for {key}")
    cells = m.group(0).split("|")
    if "**done**" in cells[6]:
        print(f"index row for {key} is already done")
    else:
        cells[5] = f" {a.date:<10} "
        cells[6] = f" {'**done**':<11} "
        text = text[:lo + m.start()] + "|".join(cells) + text[lo + m.end():]

    h = re.search(rf"^## {re.escape(key)} — .*$", text, re.M)
    if not h:
        sys.exit(f"no '## {key} — ' section")
    nxt = re.search(r"^## ", text[h.end():], re.M)
    sec_end = h.end() + nxt.start() if nxt else len(text)
    st = text.find("### Status\n\n", h.end(), sec_end)
    if st < 0:
        sys.exit(f"{key} has no '### Status' block")
    body_start = st + len("### Status\n\n")
    body_end = text.find("\n\n---", body_start, sec_end)
    body_end = body_end if body_end >= 0 else len(text[:sec_end].rstrip())
    old = text[body_start:body_end]
    if old.startswith("**Done**"):
        print(f"{key}'s status already says done")
    else:
        old_lc = re.sub(r"^(\*\*)?([A-Z])", lambda x: (x.group(1) or "") + x.group(2).lower(), old)
        new = f"**Done** ({a.date} triage). {sentence(a.note)}\n\nBefore that: {old_lc}"
        text = text[:body_start] + new + text[body_end:]
    show_or_write(path, before, text, a.dry_run)


def cmd_apply(a):
    path = Path(a.file)
    text = before = path.read_text(encoding="utf-8")
    for k, e in enumerate(json.loads(Path(a.edits).read_text(encoding="utf-8")), 1):
        if "old" in e:
            anchor, done, repl = e["old"], e["new"], None
        elif "after" in e:
            anchor, done = e["after"], e["after"] + e["insert"]
        elif "before" in e:
            anchor, done = e["before"], e["insert"] + e["before"]
        else:
            sys.exit(f"edit {k}: needs old/new, after/insert or before/insert")
        if done in text and (anchor not in text or anchor in done):
            print(f"edit {k}: already applied")
            continue
        count = text.count(anchor)
        if count != 1:
            sys.exit(f"edit {k}: anchor found {count} times, needs exactly 1: {anchor[:80]!r}")
        text = text.replace(anchor, done)
    show_or_write(path, before, text, a.dry_run)


def main():
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    sub = ap.add_subparsers(dest="cmd", required=True)
    p = sub.add_parser("render"); p.add_argument("spec")
    p.add_argument("--part", choices=["tracker", "section", "row"]); p.set_defaults(f=cmd_render)
    p = sub.add_parser("add-sprint"); p.add_argument("file"); p.add_argument("spec")
    p.add_argument("--dry-run", action="store_true"); p.set_defaults(f=cmd_add_sprint)
    p = sub.add_parser("mark-done"); p.add_argument("file"); p.add_argument("key"); p.add_argument("date")
    p.add_argument("note"); p.add_argument("--dry-run", action="store_true"); p.set_defaults(f=cmd_mark_done)
    p = sub.add_parser("apply"); p.add_argument("file"); p.add_argument("edits")
    p.add_argument("--dry-run", action="store_true"); p.set_defaults(f=cmd_apply)
    a = ap.parse_args()
    a.f(a)


if __name__ == "__main__":
    main()
