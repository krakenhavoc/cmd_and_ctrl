import json
import os
import sys
import unittest

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from post_discussion import latest_snapshot  # noqa: E402


def mark(d):
    return f"report\n<!-- cost-snapshot: {json.dumps(d)} -->"


def disc(body, *comments):
    return {"body": body, "comments": {"nodes": [{"body": c} for c in comments]}}


class LatestSnapshot(unittest.TestCase):
    def test_newest_comment_wins(self):
        d = disc(mark({"date": "a"}), mark({"date": "b"}), "chatter", mark({"date": "c"}), "more chatter")
        self.assertEqual(latest_snapshot(d)["date"], "c")

    def test_falls_back_to_body(self):
        self.assertEqual(latest_snapshot(disc(mark({"date": "a"}), "hi"))["date"], "a")

    def test_none(self):
        self.assertIsNone(latest_snapshot(disc("nothing")))


if __name__ == "__main__":
    unittest.main()
