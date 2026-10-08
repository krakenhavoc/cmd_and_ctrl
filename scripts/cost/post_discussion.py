#!/usr/bin/env python3
"""Publish the cost report to the "Project cost tracker" GitHub Discussion.

The discussion body is always the latest report. With --comment, it also
adds a comment saying what moved since the body was last written, which is
what notifies people subscribed to the thread. The previous numbers ride
in the body itself as an HTML comment, so there is no other state to keep.

    python3 scripts/cost/post_discussion.py              # create or update the body
    python3 scripts/cost/post_discussion.py --comment    # ...and post the change since last time
    python3 scripts/cost/post_discussion.py --dry-run    # print what would be posted
"""
import argparse
import json
import os
import re
import subprocess
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from cost_report import build_report, markdown, money  # noqa: E402

TITLE = "Project cost tracker"
SNAPSHOT = re.compile(r"<!-- cost-snapshot: (\{.*?\}) -->")


def graphql(query, **variables):
    cmd = ["gh", "api", "graphql", "-f", f"query={query}"]
    for k, v in variables.items():
        cmd += ["-F" if isinstance(v, int) else "-f", f"{k}={v}"]
    out = subprocess.run(cmd, capture_output=True, text=True)
    if out.returncode != 0:
        sys.exit(f"gh api graphql failed: {out.stderr.strip()}")
    return json.loads(out.stdout)["data"]


def snapshot(r):
    L = r["repo"]["lines"]
    return {
        "date": r["date"],
        "human_usd": round(r["human"]["mid"]["usd"]),
        "api_usd": round(r["ai"]["total_usd"]),
        "subs_usd": sum(v["usd"] for v in r["subscriptions"].values()),
        "merged_prs": (r["github"] or {}).get("merged_prs", 0),
        "closed_issues": (r["github"] or {}).get("closed_issues", 0),
        "code_lines": L["server"] + L["cards"] + L["client"],
        "test_lines": L["server_test"] + L["cards_test"] + L["client_test"],
        "cards": r["repo"]["cards"],
    }


def change_comment(old, new):
    def d(key, fmt):
        delta = new[key] - old.get(key, 0)
        sign = "+" if delta >= 0 else "−"
        return f"{fmt(new[key])} ({sign}{fmt(abs(delta))})"
    count = "{:,}".format
    return "\n".join([
        f"**{old.get('date', '?')} → {new['date']}**", "",
        f"- Human-team estimate: {d('human_usd', money)}",
        f"- Claude API list price: {d('api_usd', money)}",
        f"- Subscriptions paid: {d('subs_usd', money)}",
        f"- Merged PRs: {d('merged_prs', count)} · closed issues: {d('closed_issues', count)}",
        f"- Code: {d('code_lines', count)} lines · tests: {d('test_lines', count)} · cards: {d('cards', count)}",
    ]) + "\n"


def main():
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--repo", default=os.environ.get("GITHUB_REPOSITORY", "krakenhavoc/cmd_and_ctrl"))
    ap.add_argument("--category", default="General")
    ap.add_argument("--comment", action="store_true", help="also comment with the change since the last update")
    ap.add_argument("--dry-run", action="store_true")
    args = ap.parse_args()

    report = build_report()
    snap = snapshot(report)
    body = (markdown(report)
            + "\n---\n_Updated weekly by `.github/workflows/cost-report.yml`. "
              "How the numbers are made and how to add your own usage: `scripts/cost/README.md`._\n"
            + f"\n<!-- cost-snapshot: {json.dumps(snap, sort_keys=True)} -->\n")

    owner, name = args.repo.split("/")
    data = graphql("""query($owner:String!,$name:String!){repository(owner:$owner,name:$name){
        id discussionCategories(first:25){nodes{id name}}
        discussions(first:100,orderBy:{field:CREATED_AT,direction:DESC}){nodes{id title body url}}}}""",
                   owner=owner, name=name)["repository"]
    existing = next((n for n in data["discussions"]["nodes"] if n["title"] == TITLE), None)

    old = None
    if existing:
        m = SNAPSHOT.search(existing["body"] or "")
        old = json.loads(m.group(1)) if m else None
    comment = change_comment(old, snap) if (args.comment and old) else None

    if args.dry_run:
        print(body)
        if comment:
            print("--- comment ---\n" + comment)
        return

    if existing:
        graphql("mutation($id:ID!,$body:String!){updateDiscussion(input:{discussionId:$id,body:$body}){discussion{id}}}",
                id=existing["id"], body=body)
        url = existing["url"]
        if comment:
            graphql("mutation($id:ID!,$body:String!){addDiscussionComment(input:{discussionId:$id,body:$body}){comment{id}}}",
                    id=existing["id"], body=comment)
    else:
        cat = next((c["id"] for c in data["discussionCategories"]["nodes"] if c["name"] == args.category), None)
        if cat is None:
            sys.exit(f"no discussion category named {args.category!r}")
        url = graphql("""mutation($repo:ID!,$cat:ID!,$title:String!,$body:String!){
            createDiscussion(input:{repositoryId:$repo,categoryId:$cat,title:$title,body:$body}){discussion{url}}}""",
                      repo=data["id"], cat=cat, title=TITLE, body=body)["createDiscussion"]["discussion"]["url"]
    print(url)


if __name__ == "__main__":
    main()
