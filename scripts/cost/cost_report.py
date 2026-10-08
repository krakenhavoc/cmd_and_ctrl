#!/usr/bin/env python3
"""What has cmd_and_ctrl cost to build? Three answers, one report.

1. Human developers: what a team would have charged to write this codebase,
   estimated two independent ways (by component, and by merged PR) plus the
   textbook COCOMO figure for reference.
2. Claude API list price: the tokens actually used, priced as if paid per
   token. Measured from each person's ledger (scripts/cost/claude_usage.py);
   a contributor without a ledger is estimated from the measured cost per
   merged PR.
3. Subscriptions: what was actually paid, from the plan history in
   config.json.

    python3 scripts/cost/cost_report.py              # markdown to stdout
    python3 scripts/cost/cost_report.py --json       # the same numbers as JSON
    python3 scripts/cost/cost_report.py --ledger path/to/claude-usage-luke.json ...

Measures the checked-out working tree. Needs git; `gh` (authenticated) adds
PR and issue counts, which the per-PR estimates depend on.
"""
import argparse
import datetime as dt
import glob
import json
import os
import re
import subprocess
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from claude_usage import load_config, summarize  # noqa: E402

HERE = os.path.dirname(os.path.abspath(__file__))
REPO = os.path.dirname(os.path.dirname(HERE))
SCENARIOS = ("low", "mid", "high")


def sh(*cmd):
    try:
        return subprocess.run(cmd, cwd=REPO, capture_output=True, text=True, check=True).stdout
    except (OSError, subprocess.CalledProcessError):
        return None


# ---------------------------------------------------------------- the codebase

def code_lines(path):
    """Non-blank lines that are not // or /* */ comments (# for shell/yaml)."""
    try:
        with open(os.path.join(REPO, path), encoding="utf-8", errors="replace") as fh:
            text = fh.read()
    except OSError:
        return 0
    slash = path.endswith((".go", ".ts", ".js", ".svelte", ".css"))
    n, block = 0, False
    for line in text.splitlines():
        s = line.strip()
        if not s:
            continue
        if slash:
            if block:
                block = "*/" not in s
                continue
            if s.startswith("//"):
                continue
            if s.startswith("/*"):
                block = "*/" not in s
                continue
        n += 1
    return n


def area(path):
    if path.endswith(".go"):
        test = path.endswith("_test.go")
        if path.startswith("server/internal/cards/effects/"):
            return "cards_test" if test else "cards"
        return "server_test" if test else "server"
    if path.endswith((".ts", ".js", ".svelte", ".css")):
        if "node_modules/" in path or path.endswith(".config.ts") or path.endswith(".config.js"):
            return None
        if path.startswith("tests-e2e/") or re.search(r"\.(test|spec)\.ts$", path) or "/tests/" in path:
            return "client_test"
        if path.startswith("client/"):
            return "client"
        return None
    if path.endswith(".md"):
        return "docs"
    return None


def repo_stats():
    files = (sh("git", "ls-files") or "").split("\n")
    lines = dict.fromkeys(("server", "server_test", "cards", "cards_test", "client", "client_test", "docs"), 0)
    for f in files:
        a = area(f)
        if a is not None:
            lines[a] += code_lines(f)
    effects = [f for f in files if f.startswith("server/internal/cards/effects/") and f.endswith(".go")
               and not f.endswith("_test.go")]
    oracle = set()
    for f in effects:
        try:
            with open(os.path.join(REPO, f), encoding="utf-8", errors="replace") as fh:
                oracle.update(re.findall(r'OracleID:\s+"([0-9a-f-]{36})"', fh.read()))
        except OSError:
            pass
    adrs = [f for f in files if re.match(r"docs/decisions/\d{4}-.*\.md$", f)]
    first = (sh("git", "log", "--reverse", "--format=%as") or "").split()
    return {
        "lines": lines,
        "cards": len(oracle),
        "adrs": len(adrs),
        "first_commit": first[0] if first else None,
        "commits": int((sh("git", "rev-list", "--count", "HEAD") or "0").strip() or 0),
    }


def github_stats():
    """Merged PRs (author, merge day) and closed issues, or None without gh."""
    prs = sh("gh", "pr", "list", "--state", "merged", "--limit", "10000", "--json", "author,mergedAt",
             "--jq", '.[] | "\\(.author.login) \\(.mergedAt[:10])"')
    issues = sh("gh", "issue", "list", "--state", "closed", "--limit", "10000", "--json", "number", "--jq", "length")
    if prs is None:
        return None
    merged = [tuple(line.split()) for line in prs.splitlines() if line.strip()]
    per = {}
    for login, _ in merged:
        per[login] = per.get(login, 0) + 1
    return {"merged_prs": len(merged), "merged_prs_by_author": per, "merged": merged,
            "closed_issues": int(issues.strip()) if issues else None}


# ---------------------------------------------------------------- estimates

def human_estimate(stats, gh, h):
    L = stats["lines"]
    out = {}
    for s in SCENARIOS:
        hours = {
            "engine + server": L["server"] / h["engine_prod_lines_per_day"][s] * h["hours_per_day"],
            "card catalog": stats["cards"] * h["hours_per_card"][s],
            "web client": L["client"] / h["client_prod_lines_per_day"][s] * h["hours_per_day"],
            "docs + ADRs": L["docs"] / h["doc_lines_per_day"][s] * h["hours_per_day"],
        }
        base = sum(hours.values())
        hours["project overhead"] = base * h["overhead_fraction"][s]
        total = sum(hours.values())
        by_pr = gh["merged_prs"] * h["days_per_merged_pr"][s] * h["hours_per_day"] if gh else None
        out[s] = {
            "hours_by_component": hours,
            "hours": total,
            "person_years": total / h["hours_per_person_year"],
            "usd": total * h["hourly_usd"][s],
            "hours_by_pr": by_pr,
            "usd_by_pr": by_pr * h["hourly_usd"][s] if by_pr else None,
        }
    # Basic COCOMO, organic mode, on production code only. A ceiling, not a forecast.
    ksloc = (L["server"] + L["cards"] + L["client"]) / 1000
    pm = 2.4 * ksloc ** 1.05
    hours = pm * h["hours_per_person_year"] / 12
    out["cocomo"] = {"ksloc": ksloc, "person_months": pm, "person_years": pm / 12,
                     "usd": hours * h["hourly_usd"]["mid"]}
    return out


def months_between(start, end):
    y, m = map(int, start.split("-"))
    ey, em = map(int, end.split("-"))
    n = 0
    while (y, m) <= (ey, em):
        n += 1
        y, m = (y + 1, 1) if m == 12 else (y, m + 1)
    return n


def subscriptions(people, today):
    this_month = today.strftime("%Y-%m")
    out = {}
    for p in people:
        total, lines = 0, []
        for plan in p.get("plans", []):
            end = plan.get("to") or this_month
            n = months_between(plan["from"], min(end, this_month))
            total += n * plan["usd_per_month"]
            lines.append(f"{plan['name']}: {n} mo x ${plan['usd_per_month']}")
        out[p["name"]] = {"usd": total, "detail": lines}
    return out


def ai_usage(cfg, ledgers, gh):
    pricing = cfg["pricing_per_mtok"]
    measured = {}
    for path in ledgers:
        with open(path, encoding="utf-8") as fh:
            led = json.load(fh)
        by_model, by_month, total, sub, unpriced = summarize(led.get("days", {}), pricing)
        days = sorted(led.get("days", {}))
        measured[led.get("person", os.path.basename(path))] = {
            "usd": total, "subagent_usd": sub, "by_month": by_month,
            "from": days[0] if days else None, "to": days[-1] if days else None,
            "output_tokens": sum(c["output"] for c in by_model.values()),
            "cache_read_tokens": sum(c["cache_read"] for c in by_model.values()),
            "requests": sum(c["requests"] for c in by_model.values()),
            "by_model": {m: round(c["cost"], 2) for m, c in by_model.items()},
        }

    # A ledger only reaches back as far as that machine's transcripts did when
    # it was first written. PRs merged before a ledger starts are unmeasured,
    # and get the measured cost per PR, as do people with no ledger at all.
    merged = (gh or {}).get("merged", [])
    people = []
    rate_usd = rate_prs = 0
    for p in cfg["people"]:
        m = measured.get(p["github"]) or measured.get(p["name"].lower())
        mine = [day for login, day in merged if login == p["github"]]
        covered = [d for d in mine if m and m["from"] and d >= m["from"]]
        if m and covered:
            rate_usd += m["usd"]
            rate_prs += len(covered)
        people.append({"name": p["name"], "github": p["github"], "merged_prs": len(mine), "measured": m,
                       "unmeasured_prs": len(mine) - len(covered)})
    per_pr = rate_usd / rate_prs if rate_prs else None
    for p in people:
        m, n = p["measured"], p["unmeasured_prs"]
        est = per_pr * n if per_pr else 0.0
        p["usd"] = (m["usd"] if m else 0.0) + est
        parts = []
        if m:
            parts.append(f"measured {m['from']} → {m['to']}: {money(m['usd'])}, {m['requests']:,} requests, "
                         f"{m['subagent_usd'] / m['usd']:.0%} in subagents" if m["usd"] else "measured: $0")
        if n and per_pr:
            parts.append(f"estimated {money(est)} for {n} merged PRs {'before that' if m else ''} at {money(per_pr)}/PR")
        p["basis"] = "; ".join(parts) or "no data"
    return {"people": people, "usd_per_merged_pr": per_pr, "total_usd": sum(p["usd"] for p in people)}


# ---------------------------------------------------------------- report

def money(x):
    if x is None:
        return "-"
    if x >= 1e6:
        return f"${x / 1e6:,.1f}M"
    if x >= 1e4:
        return f"${x / 1e3:,.0f}K"
    return f"${x:,.0f}"


def markdown(r):
    s, gh, hu, ai, subs = r["repo"], r["github"], r["human"], r["ai"], r["subscriptions"]
    L = s["lines"]
    sub_total = sum(v["usd"] for v in subs.values())
    out = [f"# cmd_and_ctrl: what it cost to build ({r['date']})", ""]
    out += ["| | Estimate |", "|---|---|",
            f"| Human dev team | **{money(hu['mid']['usd'])}** (range {money(hu['low']['usd'])}–{money(hu['high']['usd'])}), "
            f"~{hu['mid']['person_years']:.0f} person-years |",
            f"| Claude API list price | **{money(ai['total_usd'])}** |",
            f"| Subscriptions actually paid | **{money(sub_total)}** |", ""]
    out += ["## The codebase", "",
            f"- {L['server'] + L['cards']:,} lines of Go ({L['cards']:,} of them the {s['cards']:,}-card catalog), "
            f"{L['client']:,} lines of client TypeScript/Svelte",
            f"- {L['server_test'] + L['cards_test'] + L['client_test']:,} lines of tests, "
            f"{L['docs']:,} lines of docs including {s['adrs']} ADRs"]
    if gh:
        out.append(f"- {gh['merged_prs']:,} merged PRs, {gh['closed_issues'] or 0:,} closed issues, "
                   f"{s['commits']:,} commits since {s['first_commit']}")
    out += ["", "## 1. Human developers", "",
            "| Component | low | mid | high |", "|---|---:|---:|---:|"]
    for comp in hu["mid"]["hours_by_component"]:
        out.append(f"| {comp} | " + " | ".join(f"{hu[x]['hours_by_component'][comp]:,.0f} h" for x in SCENARIOS) + " |")
    out.append("| **total** | " + " | ".join(f"**{hu[x]['hours']:,.0f} h**" for x in SCENARIOS) + " |")
    out.append("| cost | " + " | ".join(money(hu[x]["usd"]) for x in SCENARIOS) + " |")
    if gh:
        out.append("| cross-check: by merged PR | " + " | ".join(money(hu[x]["usd_by_pr"]) for x in SCENARIOS) + " |")
    c = hu["cocomo"]
    out += ["", f"Textbook COCOMO (organic, {c['ksloc']:,.0f} KSLOC of production code) says "
            f"{c['person_years']:,.0f} person-years / {money(c['usd'])}. It's included for reference only: it assumes a "
            "traditional team and counts lines, and AI-written code has more lines per feature.", ""]
    out += ["## 2. Claude API list price", "",
            "| Person | API-equivalent | Basis |", "|---|---:|---|"]
    for p in ai["people"]:
        out.append(f"| {p['name']} | {money(p['usd'])} | {p['basis'].replace('  ', ' ')} |")
    out.append(f"| **total** | **{money(ai['total_usd'])}** | |")
    out += ["", "## 3. Subscriptions", "", "| Person | Paid | Plans |", "|---|---:|---|"]
    for name, v in subs.items():
        out.append(f"| {name} | {money(v['usd'])} | {'; '.join(v['detail'])} |")
    if sub_total:
        out += ["", f"API-equivalent per subscription dollar: **{ai['total_usd'] / sub_total:,.0f}x**. "
                f"Human-team estimate per subscription dollar: **{hu['mid']['usd'] / sub_total:,.0f}x**."]
    return "\n".join(out) + "\n"


def build_report(ledgers=None):
    cfg = load_config()
    if ledgers is None:
        ledgers = sorted(glob.glob(os.path.join(REPO, "data", "cost", "claude-usage-*.json")))
    stats = repo_stats()
    gh = github_stats()
    today = dt.date.today()
    return {
        "date": today.isoformat(),
        "repo": stats,
        "github": gh,
        "human": human_estimate(stats, gh, cfg["human"]),
        "ai": ai_usage(cfg, ledgers, gh),
        "subscriptions": subscriptions(cfg["people"], today),
    }


def main():
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--ledger", action="append", default=None,
                    help="usage ledger(s); default data/cost/claude-usage-*.json")
    ap.add_argument("--json", action="store_true")
    args = ap.parse_args()
    report = build_report(args.ledger)
    if args.json:
        json.dump(report, sys.stdout, indent=1, default=str)
        print()
    else:
        sys.stdout.write(markdown(report))


if __name__ == "__main__":
    main()
