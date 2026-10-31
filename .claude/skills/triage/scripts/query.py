#!/usr/bin/env -S python3 -I
"""Read-only views over a triage snapshot (take one with snapshot.sh first).

  query.py lines                  one line per open issue
  query.py gaps                   issues still missing a milestone, a type, one tier or one priority
  query.py urgent                 P0, P1 and player-report issues, worst first
  query.py show <n>               an issue's labels, body and comments
  query.py recent [--days 7]      comments posted in the last N days
  query.py claims [--days 7]      claimed issues, and whether each claim still holds
  query.py questions              parked owner questions, and whether anyone has replied since
  query.py evidence <n> [--live]  PRs naming an issue: directive or mention, merged, on develop, on main
  query.py done                   open issues whose closing PR is merged into develop: close candidates
  query.py milestones             open milestones with their counts

The snapshot directory is $TRIAGE_DIR, or $CLAUDE_JOB_DIR/tmp/triage. `evidence`
and `done` also read git (origin/develop, origin/main); snapshot.sh fetched it.
Nothing here writes to GitHub.
"""

import argparse
import datetime as dt
import json
import os
import re
import subprocess
import sys
from pathlib import Path

ROOT = Path(
    subprocess.run(
        ["git", "-C", str(Path(__file__).resolve().parent), "rev-parse", "--show-toplevel"],
        capture_output=True, text=True, check=True,
    ).stdout.strip()
)

# `question` counts as a type here, as it did in the 2026-10-30 audit.
TYPES = {"bug", "enhancement", "documentation", "question", "tech-debt", "seam"}
PRIORITIES = ["P0", "P1", "P2", "P3"]
CARDS_BATCH = re.compile(r"^Cards batch \d+:")  # milestone-less by design (issue-triage.yml)
TRACKER = re.compile(r"^S\d+(\.\d+)? — ")  # a sprint tracker is titled like its milestone
CLAIM = re.compile(
    r"(?i)^\W*claim(ed|ing)?\b|\b(picking|picked) (this|it) up\b|\bworking on (this|it)\b"
    r"|\btaking (this|it)( on)?\b|\bi'?ll take (this|it)\b"
)

# The promotion closer (.github/workflows/main-promotion-issue-close.yml) acts
# only on directive lines inside a PR body's "## Issues" section. These mirror
# its parser, so "has a closing directive" here means what it means there.
KW = r"(?:close(?:s|d)?|fix(?:es|ed)?|resolve(?:s|d)?)"
DIRECTIVE = re.compile(
    rf"^\s*(?:[-*]\s*)?({KW}(?:\s*:\s*|\s+)#[1-9]\d*(?:\s*,\s*(?:{KW}(?:\s*:\s*|\s+))?#[1-9]\d*)*)"
    r"(?=$|[\s,.;:·—–()])",
    re.I,
)


def snapshot_dir() -> Path:
    d = os.environ.get("TRIAGE_DIR") or (
        os.path.join(os.environ["CLAUDE_JOB_DIR"], "tmp", "triage") if os.environ.get("CLAUDE_JOB_DIR") else ""
    )
    if not d or not Path(d, "issues.json").exists():
        sys.exit("no snapshot: run snapshot.sh first (TRIAGE_DIR or CLAUDE_JOB_DIR picks the directory)")
    return Path(d)


def load(name):
    return json.loads((snapshot_dir() / name).read_text(encoding="utf-8"))


def issue_detail(n: int):
    p = snapshot_dir() / "iss" / f"{n}.json"
    return json.loads(p.read_text(encoding="utf-8")) if p.exists() else None


def labels(issue) -> list:
    return [l["name"] for l in issue.get("labels") or []]


def short_milestone(issue) -> str:
    m = issue.get("milestone")
    return m["title"].split(" — ")[0] if m else "-"


def one_line(text: str, width: int = 160) -> str:
    text = " ".join((text or "").split())
    return text if len(text) <= width else text[: width - 1] + "…"


def cutoff(days: int) -> str:
    t = dt.datetime.now(dt.timezone.utc) - dt.timedelta(days=days)
    return t.strftime("%Y-%m-%dT%H:%M:%SZ")


# ---- pull requests and git -------------------------------------------------

def issues_section(body: str) -> list:
    out, active, fenced, comment = [], False, False, False
    fence_char, fence_len = "", 0
    for line in (body or "").replace("\r\n", "\n").split("\n"):
        if fenced:
            if not (line.startswith("    ") or line.startswith("\t")):
                text = line.lstrip(" ")
                n = len(text) - len(text.lstrip(fence_char))
                if n >= fence_len and not text[n:].strip():
                    fenced = False
            continue
        if comment:
            if "-->" in line:
                comment = False
            continue
        if line.startswith("    ") or line.startswith("\t"):
            continue
        m = re.match(r"^ *(`{3,}|~{3,})", line)
        if m:
            fenced, fence_char, fence_len = True, m.group(1)[0], len(m.group(1))
            continue
        if "<!--" in line:
            comment = "-->" not in line
            continue
        if re.match(r"^\s*>", line):
            continue
        if re.match(r"^ *##\s+issues\s*$", line, re.I):
            active = True
            continue
        if active and re.match(r"^ *##\s+", line):
            break
        if active:
            out.append(line)
    return out


def closes(pr) -> set:
    found = set()
    for line in issues_section(pr.get("body") or ""):
        m = DIRECTIVE.match(line)
        if m:
            found.update(int(x) for x in re.findall(r"#([1-9]\d*)", m.group(1)))
    return found


def names(pr, n: int) -> bool:
    pat = re.compile(rf"(?<![\w/])#{n}(?!\d)")
    return bool(pat.search(pr.get("title") or "") or pat.search(pr.get("body") or ""))


_ancestor_cache = {}


def on_branch(sha: str, ref: str) -> str:
    if not sha:
        return "?"
    key = (sha, ref)
    if key not in _ancestor_cache:
        r = subprocess.run(["git", "-C", str(ROOT), "merge-base", "--is-ancestor", sha, ref], capture_output=True)
        _ancestor_cache[key] = {0: "yes", 1: "no"}.get(r.returncode, "?")
    return _ancestor_cache[key]


def merge_sha(pr) -> str:
    return ((pr.get("mergeCommit") or {}).get("oid")) or ""


def is_promotion(pr) -> bool:
    return pr.get("headRefName") == "develop" and pr.get("baseRefName") == "main"


def prs_naming(n: int, prs: list) -> list:
    return [p for p in prs if not is_promotion(p) and names(p, n)]


def live_prs(n: int) -> list:
    repo = os.environ.get("TRIAGE_REPO") or subprocess.run(
        ["gh", "repo", "view", "--json", "nameWithOwner", "--jq", ".nameWithOwner"],
        cwd=ROOT, capture_output=True, text=True, check=True,
    ).stdout.strip()
    out = subprocess.run(
        ["gh", "pr", "list", "-R", repo, "--state", "all", "--limit", "100", "--search", str(n), "--json",
         "number,title,state,isDraft,mergedAt,baseRefName,headRefName,author,body,mergeCommit,createdAt"],
        capture_output=True, text=True, check=True,
    ).stdout
    return json.loads(out)


def describe_pr(pr, n: int) -> str:
    kind = "closes" if n in closes(pr) else "mentions"
    state = pr["state"].lower() + (" draft" if pr.get("isDraft") else "")
    line = f"PR #{pr['number']} {state} {kind} base={pr.get('baseRefName')}"
    if pr.get("mergedAt"):
        sha = merge_sha(pr)
        line += (f" merged {pr['mergedAt'][:10]} as {sha[:9]}"
                 f" develop={on_branch(sha, 'origin/develop')} main={on_branch(sha, 'origin/main')}")
    return f"{line} @{(pr.get('author') or {}).get('login', '?')}: {one_line(pr.get('title'), 90)}"


# ---- claims ----------------------------------------------------------------

def claim_state(issue, prs: list, days: int):
    """Returns (status, detail), or None when nothing claims the issue."""
    n = issue["number"]
    detail = issue_detail(n) or issue
    naming = prs_naming(n, prs)
    merged = [p for p in naming if p.get("mergedAt") and n in closes(p)
              and on_branch(merge_sha(p), "origin/develop") == "yes"]
    open_prs = [p for p in naming if p["state"] == "OPEN"]
    claims = [c for c in detail.get("comments") or [] if CLAIM.search(c.get("body") or "")]
    last = claims[-1] if claims else None
    assignees = [a["login"] for a in issue.get("assignees") or []]
    in_progress = "in-progress" in labels(issue)
    if not (last or assignees or in_progress or open_prs or merged):
        return None
    bits = []
    if assignees:
        bits.append("assigned " + ",".join(assignees))
    if in_progress:
        bits.append("in-progress")
    if last:
        bits.append(f"claim by {last['author']['login']} {last['createdAt'][:10]}")
    if open_prs:
        bits.append("open " + ",".join(f"#{p['number']}" for p in open_prs))
    if merged:
        status = "MERGED: close with evidence (" + ",".join(f"#{p['number']}" for p in merged) + ")"
    elif open_prs:
        status = "ACTIVE: open PR, hands off"
    elif last and last["createdAt"] >= cutoff(days):
        status = f"ACTIVE: claimed within {days} days, hands off"
    elif last:
        status = f"STALE: claim older than {days} days, no open PR"
    else:
        status = "UNDATED: assignee or label but no claim comment; read the timeline"
    return status, "; ".join(bits)


# ---- commands --------------------------------------------------------------

def cmd_lines(a):
    for i in sorted(load("issues.json"), key=lambda i: i["number"]):
        print(f"#{i['number']} [{short_milestone(i)}] {{{','.join(labels(i))}}} "
              f"@{i['author']['login']} {i['createdAt'][:10]} {one_line(i['title'], 110)}")


def cmd_gaps(a):
    rows = 0
    for i in sorted(load("issues.json"), key=lambda i: i["number"]):
        ls = labels(i)
        why = []
        if i.get("milestone") is None and not CARDS_BATCH.match(i["title"]):
            why.append("no milestone")
        tiers = [l for l in ls if l.startswith("tier:")]
        if len(tiers) != 1:
            why.append(f"{len(tiers)} tiers")
        prios = [l for l in ls if l in PRIORITIES]
        if len(prios) != 1:
            why.append(f"{len(prios)} priorities")
        if not TYPES & set(ls):
            why.append("no type")
        if "needs-triage" in ls:
            why.append("needs-triage")
        if why:
            rows += 1
            print(f"#{i['number']} {', '.join(why)} | {{{','.join(ls)}}} {one_line(i['title'], 90)}")
    print(f"{rows} issue(s) with gaps", file=sys.stderr)


def cmd_urgent(a):
    def rank(i):
        ls = labels(i)
        p = next((PRIORITIES.index(l) for l in ls if l in PRIORITIES), None)
        if p is not None and p <= 1:
            return (p, i["number"])
        return (2 if p is None else 3 + p, i["number"])
    hits = [i for i in load("issues.json")
            if {"P0", "P1", "player-report"} & set(labels(i))]
    for i in sorted(hits, key=rank):
        ls = labels(i)
        prio = next((l for l in ls if l in PRIORITIES), "no priority")
        tag = " player-report" if "player-report" in ls else ""
        print(f"#{i['number']} {prio}{tag} [{short_milestone(i)}] {one_line(i['title'], 110)}")


def cmd_show(a):
    d = issue_detail(a.n)
    if not d:
        sys.exit(f"#{a.n} is not in the snapshot (closed, or newer than it)")
    print(f"#{d['number']} {d['title']}")
    print(f"state {d.get('state', '?')} · milestone {(d.get('milestone') or {}).get('title', '-')} · "
          f"labels {','.join(labels(d)) or '-'} · assignees "
          f"{','.join(x['login'] for x in d.get('assignees') or []) or '-'} · "
          f"by @{d['author']['login']} {d['createdAt'][:10]}")
    print("\n" + (d.get("body") or "(no body)").strip())
    for c in d.get("comments") or []:
        print(f"\n--- @{c['author']['login']} {c['createdAt'][:16].replace('T', ' ')}")
        print((c.get("body") or "").strip())


def cmd_recent(a):
    since = cutoff(a.days)
    for i in sorted(load("issues.json"), key=lambda i: i["number"]):
        d = issue_detail(i["number"]) or {}
        for c in d.get("comments") or []:
            if c["createdAt"] >= since:
                print(f"#{i['number']} @{c['author']['login']} {c['createdAt'][:10]}: {one_line(c.get('body'))}")


def cmd_claims(a):
    prs = load("prs.json")
    for i in sorted(load("issues.json"), key=lambda i: i["number"]):
        if TRACKER.match(i["title"]):
            continue  # every sprint PR names its tracker; that isn't a claim
        s = claim_state(i, prs, a.days)
        if s:
            print(f"#{i['number']} {s[0]} | {s[1]} | {one_line(i['title'], 80)}")


def cmd_questions(a):
    for i in sorted(load("issues.json"), key=lambda i: i["number"]):
        if "question" not in labels(i):
            continue
        comments = (issue_detail(i["number"]) or {}).get("comments") or []
        parked = [k for k, c in enumerate(comments) if "Owner decision needed" in (c.get("body") or "")]
        if not parked:
            print(f"#{i['number']} question label, no parked-question comment | {one_line(i['title'], 90)}")
            continue
        k = parked[-1]
        after = comments[k + 1:]
        reply = (f"{len(after)} later comment(s), last @{after[-1]['author']['login']} {after[-1]['createdAt'][:10]}"
                 if after else "no reply yet")
        print(f"#{i['number']} parked {comments[k]['createdAt'][:10]}, {reply} | {one_line(i['title'], 90)}")


def cmd_evidence(a):
    prs = live_prs(a.n) if a.live else load("prs.json")
    naming = sorted(prs_naming(a.n, prs), key=lambda p: p["number"])
    if not naming and not a.live:
        print(f"no PR in the snapshot window names #{a.n}; searching GitHub")
        naming = sorted(prs_naming(a.n, live_prs(a.n)), key=lambda p: p["number"])
    for p in naming:
        print(describe_pr(p, a.n))
    if not naming:
        print(f"no PR names #{a.n}")
    log = subprocess.run(
        ["git", "-C", str(ROOT), "log", "origin/develop", "-E", f"--grep=#{a.n}([^0-9]|$)",
         "--format=%h %ad %s", "--date=short", "-n", "15"],
        capture_output=True, text=True,
    ).stdout.strip()
    print(f"commits on origin/develop naming #{a.n}:" + ("\n" + log if log else " none"))
    print("A closing directive is a lead, not proof: confirm the change on origin/develop and that it covers the issue.")


def cmd_done(a):
    prs = load("prs.json")
    by_issue = {}
    for p in prs:
        if not p.get("mergedAt") or is_promotion(p):
            continue
        for n in closes(p):
            by_issue.setdefault(n, []).append(p)
    for i in sorted(load("issues.json"), key=lambda i: i["number"]):
        hits = [p for p in by_issue.get(i["number"], []) if on_branch(merge_sha(p), "origin/develop") == "yes"]
        if not hits:
            continue
        refs = ", ".join(f"#{p['number']} as {merge_sha(p)[:9]} {p['mergedAt'][:10]} main={on_branch(merge_sha(p), 'origin/main')}"
                         for p in hits)
        claim = claim_state(i, prs, 7)
        tag = f" | {claim[1]}" if claim and claim[1] else ""
        print(f"#{i['number']} closed by {refs}{tag} | {one_line(i['title'], 80)}")


def cmd_milestones(a):
    for m in sorted(load("milestones.json"), key=lambda m: m["number"]):
        if m["state"] == "open":
            print(f"{m['number']}\t{m['open_issues']} open / {m['closed_issues']} closed\t{m['title']}")


def main():
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    sub = ap.add_subparsers(dest="cmd", required=True)
    sub.add_parser("lines").set_defaults(f=cmd_lines)
    sub.add_parser("gaps").set_defaults(f=cmd_gaps)
    sub.add_parser("urgent").set_defaults(f=cmd_urgent)
    p = sub.add_parser("show"); p.add_argument("n", type=lambda s: int(s.lstrip("#"))); p.set_defaults(f=cmd_show)
    p = sub.add_parser("recent"); p.add_argument("--days", type=int, default=7); p.set_defaults(f=cmd_recent)
    p = sub.add_parser("claims"); p.add_argument("--days", type=int, default=7); p.set_defaults(f=cmd_claims)
    sub.add_parser("questions").set_defaults(f=cmd_questions)
    p = sub.add_parser("evidence"); p.add_argument("n", type=lambda s: int(s.lstrip("#")))
    p.add_argument("--live", action="store_true", help="search GitHub instead of the snapshot")
    p.set_defaults(f=cmd_evidence)
    sub.add_parser("done").set_defaults(f=cmd_done)
    sub.add_parser("milestones").set_defaults(f=cmd_milestones)
    a = ap.parse_args()
    a.f(a)


if __name__ == "__main__":
    main()
