# ADR 0220 — Pods: groups of players for shared stats

**Status:** Accepted · 2026-11-05 · outside the sprint's scope (owner request; see "Sprint" below)
**Issue:** [#4880](https://github.com/krakenhavoc/cmd_and_ctrl/issues/4880) (the stats epic)
**Owner decisions:** Ian, in ADR 0217's [owner answers](0217-advanced-player-stats-per-deck-and-overall.md#owner-answers-to-the-open-questions-ian-2026-11-04) 5 and 7 and in the brief for this ADR (2026-11-05). They are recorded below as [Owner decisions](#owner-decisions) and are **decided**. Everything else here is a call made in this ADR, listed under [Calls made here](#calls-made-here), and can be overturned in review.
**Numbering:** checked on 2026-11-05 after `git fetch --all --prune`. Every ADR file ever added on any remote head (`git log --remotes=origin --name-only -- docs/decisions/`, a superset of the loop in [README.md](README.md#picking-a-number)) shows 0218 and 0219 taken on `develop` and nothing above 0219 anywhere. The heads checked are listed in the PR. This is **0220**.
**Builds on:**
- [ADR 0217](0217-advanced-player-stats-per-deck-and-overall.md) §6, §8, §9 and slice 7: what the stats side needs from a pod, `stats_prefs.share_pods`, and the pod stats routes. This ADR is the "pod ADR" it requires.
- [ADR 0213](0213-friends-recently-played-with-and-live-presence.md): mutual friends (§2), personal blocks (§3), the `social_prefs.requests` setting, and its limits.
- [ADR 0149](0149-privacy-notice-and-account-deletion.md): the tombstone, `accountdelete.Inventory` and `Delete`, the privacy page, `PRIVACY_SCHEMA_VERSION`.
- [ADR 0148](0148-capacity-limits-for-a-public-launch.md): the beta's size, about 120 players on one VM.
- [ADR 0075](0075-table-settings-and-host-controls.md) §2.6: the name filter (`internal/moderation`, `name_not_allowed`).
- [ADR 0156](0156-a-front-door-site-shell-front-page-and-a-discord-gate.md) §3: the grouped site nav.

---

## Context

ADR 0217 lets a person see their own stats and share them by link. Its owner decision 8 also wants a **pod**: a group of players who see each other's stats and a leaderboard. ADR 0217 §6 left the group itself to this ADR: who makes one, how people join, the name, leaving, limits and deletion.

What the code gives us to work with, read on 2026-11-05:

- **No group concept exists.** Nothing in the schema names a set of people.
- **The stats side is ready for it.** `stats_prefs.share_pods` (migration 0030) exists and is 0. The share link (`lobby/stats_share.go`) already mints a 128-bit token and keeps only its SHA-256, which is the pattern a pod invite link can copy.
- **Friends and blocks are designed but not built.** ADR 0213 is accepted, but its slice 3 (`friendships`, `friend_requests`, `user_blocks`, `social_prefs`) has not landed (#3313 is open). Pods need both: friends to invite from, and blocks to respect.
- **Who played with whom is a query.** `GET /me/tablemates` is a self-join of `seats` (ADR 0051 decision 8). The same join can find a group that keeps sitting together.
- **The name filter exists.** `moderation.Name` refuses a listed name with a 400 and the code `name_not_allowed`, for guest names, Discord display names, table names and scheduled-game names.
- **Reports are per game.** `player_reports` is filed from a seat at a table (ADR 0075 §2.6). There is no report for anything outside a game.
- **Account deletion is a tombstone.** The `users` row stays (ADR 0149 §2), so `ON DELETE CASCADE` from `users` never fires. `accountdelete.Delete` must remove a person's rows itself.

## Decision

### 1. What a pod is

- **A pod is a named set of accounts**, with one **owner**. It has no other content: no chat, no description, no picture.
- **Who can be a member:** a signed-in person. Guests, bots, agent seats and the admin token have no account and cannot join (ADR 0217 §6).
- **A pod is private.** Only its members can read it. Every route answers **404** to anyone else, so a pod id cannot be probed. There is no pod search and no public list.
- **Ids** are random UUIDs, like `games.id`.

### 2. Roles

There are two roles, **owner** and **member**. One owner is the cheapest answer to "who decides" in a group of friends.

| Action | Owner | Member |
|---|---|---|
| See the pod, its members and who is sharing | yes | yes |
| Invite a friend | yes | yes |
| Make, copy, rotate or turn off the invite link | yes | no |
| Rename the pod | yes | no |
| Remove a member | yes | no |
| Hand the pod to another member | yes | no |
| Delete the pod | yes | no |
| Leave | yes | yes |

- **The creator is the owner.**
- **Handing over:** the owner can make any member the owner. They stay a member.
- **When the owner leaves or deletes their account,** the member who joined first becomes the owner. Nobody is asked. The pod's page shows the new owner.
- **When the last member leaves,** the pod is deleted.

### 3. Joining

There are three ways in. Each ends with the person's own yes.

**a. A friend invite.**

- Any member can invite one of **their own friends** (ADR 0213 §2). It is a pending invite until the friend answers.
- **Accept** makes them a member. **Decline** is silent, like ADR 0213's Ignore: the invite leaves their list, and the inviter still sees it as pending until it expires.
- Invites expire after **30 days**, as friend requests do.
- An invite carries no text, so it adds nothing to moderate.

**b. The invite link.**

- The owner mints it. Like the stats share link (ADR 0217 §6): a 128-bit token in base64url, and only its SHA-256 is stored, in `pods.invite_hash`. The token is shown once, when it is made.
- It opens `#/pods/join/<token>`. A signed-out visitor sees only "Sign in with Discord to join this pod". Nothing about the pod is shown before sign-in.
- Signed in, the page shows the pod's name, the owner's display name and the member count, with **Join** and **Not now**. Clicking Join is the yes.
- A link **expires after 7 days**. The owner can rotate it (the old one stops at once) or turn it off.
- It has no use count. The size cap (§5) bounds it.
- An unknown, rotated, expired or turned-off link, a full pod, and a join refused for a block (§6) all get the same answer: "This invite link can't be used." A probe learns nothing.

**c. A suggestion** (§4). Accepting one makes a pod and sends ordinary invites to the others, who each accept or decline.

**Leaving** is one click, from either the pod page or the list, with no confirm beyond "Leave <name>?". It is silent: nobody is told, and the person just stops being listed.

**Removing a member** is the owner's. The Remove dialog has a checkbox, on by default, **"Also make a new invite link"**, because a removed person who still holds the link could rejoin. Removal stores nothing about the person.

### 4. Suggested pods

C2 may suggest a pod to people who often play together (owner answer 5). A suggestion forms nothing on its own.

- **The signal:** the same **3 or 4 signed-in people** sat together at **3 or more ended games in the last 30 days**.
  - "The same people" is the exact set of signed-in, non-bot, non-agent seats at each game. Guests and bots at the table do not stop a game counting, and they are never part of the set.
  - Two people are not suggested a pod. A pair is what friends are for.
- **Who sees it:** each person in the set, on `#/pods`. The card names the others and says "You've played 4 games together in the last 30 days." Each of them already sees the others in their Played with list (ADR 0213 §1), so the card tells nobody anything new.
- **Who is left out:** no suggestion is made for a set where
  - any two of them have a block between them (§6);
  - anyone has set "Who can send me friend requests" to *nobody* (ADR 0213 §2), so they never get an invite from someone who is not their friend;
  - anyone's account is deleted;
  - a pod already holds all of them.
- **Worked out when read.** `GET /me/pods` computes the caller's suggestions from `seats` and `games`. Nothing is stored for a suggestion except a dismissal.
- **Make this pod:** asks for a name. The person becomes the owner, and the others get invites as in §3a. This is the one way to invite someone who is not your friend, and only the people in the suggestion.
- **Not now:** stores a dismissal, keyed by the caller and a hash of the set's sorted account ids. It hides that suggestion from the caller for **180 days**, then the suggestion may come back if the group still meets the signal. It hides it from nobody else.

### 5. Limits

| Limit | Value | Why |
|---|---|---|
| Members in a pod | **12** | three tables of four; a playgroup, not a community |
| Pods a person is in (owned or not) | **10** | enough for several groups; bounds every pod read |
| Pending invites in a pod | **20** | |
| Invites sent | **10 an hour** per person, counted in memory | as ADR 0213's friend requests |
| Making a pod | a burst of **3**, then **1 an hour** per person | |
| Invite expiry | **30 days** | as friend requests |
| Invite link expiry | **7 days** | a posted link stops working on its own |
| Dismissed suggestion | **180 days** | |
| Name | **1 to 40 characters** after trimming | fits a header and a list row |

A refusal at a cap names nobody. Accepting an invite to a full pod says "This pod is full". A link join to a full pod gets §3b's one answer. Anyone already in 10 pods is told "You're in 10 pods. Leave one to join another."

**Against ADR 0148.** At its target of 120 players, the most this can store is 1,200 membership rows and no more pods than that, a few hundred KB. A pod read (§7) touches at most 12 members' games. The routes use the existing per-caller limiter: the writes share `statsPrefsLimit`'s shape (burst 5, 1 a second), and the reads share `advancedStatsLimit`'s (burst 10, 2 a second). There is no cache until a measurement asks for one.

### 6. Blocks (ADR 0213 §3)

- **A blocked person cannot get in.** A friend invite between them is impossible, because a block ends the friendship. A link join is refused if the joiner and **any** member have a block between them, in either direction, with the same answer as a dead link (§3b). A suggestion is never made for the set.
- **A block between two members, made after both joined, removes neither.** A block is personal, and removing one of them would tell them. Instead each one disappears from the other's view of the pod:
  - from the member list;
  - from the leaderboard (the other rows keep their numbers, and their ranks close up);
  - from head-to-head;
  - and their name is never shown to the other on the pod's games.
- **The pod's numbers do not change.** "Pod games" stay games where two or more sharing members sat (ADR 0217 §6), counted over every sharing member, so one member's block does not change another member's record. A block hides a person; it does not rewrite the group's history.
- **The owner can still remove** a member they have blocked. The owner's member list keeps that one row, marked **Blocked**, with Remove and nothing else about them. A member who has blocked the owner is hidden from the owner like anyone else, and can leave.

### 7. Stats and the leaderboard (ADR 0217 §6)

This ADR does not change the stats side. It confirms:

- **Joining is not consent to share.** `stats_prefs.share_pods` is. It is **one switch for every pod** the person is in, as ADR 0217 built it. A member who has not switched it on shows as **Not sharing**, with no numbers.
- **Where the switch is:** Settings → Stats, and the pod page ("Share my stats with my pods"). The pod page says in one sentence that it covers all of the person's pods.
- **What a member who is not sharing still shows** to the pod: their display name, avatar and that they are a member. That is what a member list is.
- **`GET /pods/{id}/stats`** and **`GET /pods/{id}/leaderboard`**, members only, as ADR 0217 §6 defines them, with §6's block filter above.
- **The leaderboard covers pod games only** (owner answer 7). There is no "All games" scope. It ranks sharing members with at least **5 counted games** in the range, by win rate, then average place, then games. Every row shows its count.
- **History counts.** Pod games include games from before the pod was made, because the switch shares the person's record, not just what comes after. The range filter (all time, 90 days, 30 days) is how a pod looks at a recent stretch.

### 8. The name

- **Free text, 1 to 40 characters** after trimming, one line, no control characters.
- **The same filter as every other name:** `moderation.Name`. A listed word is refused with a 400 and `name_not_allowed`, without saying which word (ADR 0075 §2.6).
- **Who sees it:** members, and a signed-in person holding a live invite link or invite. Never anyone signed out.
- **No new report kind.** The audience is a handful of people who chose to be there, and each can leave. Two admin routes cover a bad name that gets through: `PUT /admin/pods/{id}/name` and `DELETE /admin/pods/{id}`, behind `requireAdmin`, which writes the `admin action` audit line. A member who needs an admin sends the pod's page link through the usual support path. If pod names turn out to be a problem, a report kind is a later, additive change.
- **Admins get no pod list** and no membership graph, as ADR 0213 keeps no friend graph. The two routes act on one pod by id.

### 9. Schema

**One migration** (the next free number when slice 2 starts; `develop` is at 0030 on 2026-11-05). Additive. No new personal data beyond who is in which pod.

```sql
CREATE TABLE pods (
    id                TEXT PRIMARY KEY,                 -- random UUID
    name              TEXT NOT NULL,                    -- §8, filtered
    owner_id          TEXT NOT NULL REFERENCES users(id),
    invite_hash       TEXT UNIQUE,                      -- SHA-256 hex of the link token, or NULL
    invite_expires_at INTEGER,                          -- unix ms, NULL when no link
    created_at        INTEGER NOT NULL,
    updated_at        INTEGER NOT NULL,
    CHECK ((invite_hash IS NULL) = (invite_expires_at IS NULL))
);

CREATE TABLE pod_members (
    pod_id    TEXT NOT NULL REFERENCES pods(id) ON DELETE CASCADE,
    user_id   TEXT NOT NULL REFERENCES users(id),
    joined_at INTEGER NOT NULL,
    PRIMARY KEY (pod_id, user_id)
);
CREATE INDEX pod_members_user ON pod_members(user_id);

CREATE TABLE pod_invites (
    pod_id      TEXT NOT NULL REFERENCES pods(id) ON DELETE CASCADE,
    user_id     TEXT NOT NULL REFERENCES users(id),     -- the invitee
    invited_by  TEXT NOT NULL REFERENCES users(id),
    created_at  INTEGER NOT NULL,
    declined_at INTEGER,                                -- silent decline, §3a
    PRIMARY KEY (pod_id, user_id)
);
CREATE INDEX pod_invites_user ON pod_invites(user_id);

CREATE TABLE pod_suggestion_dismissals (
    user_id      TEXT NOT NULL REFERENCES users(id),
    group_key    TEXT NOT NULL,                         -- SHA-256 hex of the sorted account ids
    dismissed_at INTEGER NOT NULL,
    PRIMARY KEY (user_id, group_key)
);
```

- **The owner is always a member.** The code holds it, in the same transaction as every change of owner, and a test checks it.
- **Deleting a pod** removes its members and invites by cascade.
- **Expiry:** a sweeper, run like `RunReportSweeper`, deletes invites past 30 days, dismissals past 180 days, and clears `invite_hash` past `invite_expires_at`. A read also treats an expired row as gone, so the sweep is tidiness, not a guarantee.
- **Not stored:** suggestions themselves, the invite token, who joined by which way, and anything about a person removed or a link refused.

### 10. Routes

All are behind `auth.Middleware`, and a guest, spectator or admin-token caller gets 403 ("pods need an account"). Reads and writes take the per-caller limits of §5. Non-members get 404.

| Route | Who | Does |
|---|---|---|
| `GET /me/pods` | anyone signed in | my pods (name, member count, my role), my pending invites, my suggestions |
| `POST /pods` `{name, suggestion?}` | anyone signed in | makes a pod; with `suggestion` (its group key), also invites the set (§4) |
| `GET /pods/{id}` | members | name, owner, members (display name, avatar, joined, sharing or not), pending invites; for the owner, whether a link is live and when it expires |
| `PATCH /pods/{id}` `{name}` | owner | rename |
| `DELETE /pods/{id}` | owner | delete |
| `PUT /pods/{id}/owner` `{user_id}` | owner | hand over to a member |
| `POST /pods/{id}/invites` `{user_id}` | members | invite a friend |
| `DELETE /pods/{id}/invites/{user_id}` | the inviter or the owner | cancel an invite |
| `POST /me/pod-invites/{pod_id}/accept`, `/decline` | the invitee | answer |
| `POST /pods/{id}/link`, `DELETE /pods/{id}/link` | owner | mint or rotate (answers the token once), turn off |
| `GET /pods/join/{token}`, `POST /pods/join/{token}` | anyone signed in | preview, join |
| `DELETE /pods/{id}/members/{user_id}` | owner, or the member themselves (`me`) | remove, leave |
| `POST /me/pod-suggestions/{group_key}/dismiss` | a person in the set | Not now |
| `PUT /admin/pods/{id}/name`, `DELETE /admin/pods/{id}` | admin | §8 |
| `GET /pods/{id}/stats`, `GET /pods/{id}/leaderboard` | members | ADR 0217 slice 7 |

- **Account ids** reach the client only for people the caller can already see: their friends (to invite) and their pods' members (to remove or hand over). A suggestion's group key is a hash, not a list of ids.
- **Refusals** read like ADR 0213's ("can't send this to that person"), so a block is never named to the blocked person.
- **Bans and losing the gate** (ADR 0156) leave memberships in place, as friendships stay (ADR 0213 §2). The person cannot sign in, so they cannot read the pod.
- docs/lobby.md lists every route in the slice that adds it.

### 11. The pages

Following DESIGN.md. Site pages in `Page.svelte`. Every label is `L.<key>` from `labels.ts` and listed in `docs/labels.md`.

- **`#/pods`, my pods.** Linked from the signed-in nav group that holds Stats (ADR 0156 §3), and from a Pods link on `#/stats`.
  - **Invites** first, when there are any: the pod's name, who invited you, **Accept** and **Decline**.
  - **Suggested**, when there are any: the others' names and avatars, the games count, **Make this pod** and **Not now**.
  - **Your pods:** one `.card` row each, with the name, the member count, an Owner `.tag` on your own, and a link to the pod.
  - **New pod** is the page's one primary.
  - Empty state: "A pod is a group you play with. Make one, then invite your friends."
- **`#/pods/<id>`, a pod.**
  - The name as the page title, the owner and the member count as its meta.
  - Members: avatar, display name, joined date, and a **Sharing** or **Not sharing** tag. Owner actions on a row sit in a `ContextMenu` (Remove, Make owner).
  - Your own sharing switch, with its one sentence (§7).
  - **Invite a friend** opens a friend picker built like `TablematePicker.svelte`. Friends already in the pod, already invited, or with a block against any member are not offered.
  - Owner only: **Invite link** (Copy, New link, Turn off, with its expiry date), **Rename**, **Delete pod**.
  - **Leave pod** for everyone.
  - A **Stats** link to `#/pods/<id>/stats` (ADR 0217 §8), once slice 5 lands.
- **`#/pods/join/<token>`:** the preview of §3b, or the one sign-in line.
- **Phones:** one column below 819 px (ADR 0217 owner answer 9). Screenshots at 1440, 820 and 390 px with each client slice.

### 12. Privacy (ADR 0149)

| Data | Where | Kept for | Deleting the account |
|---|---|---|---|
| `pods` (name, owner, link hash and expiry) | DB | until the pod is deleted | a pod the person owns passes to the member who joined first, or is **deleted** if they were the last. Its name stays with the pod: it belongs to the group, like a table name |
| `pod_members` | DB | until the person leaves or is removed, or the pod is deleted | the person's rows are **deleted** |
| `pod_invites` | DB | until answered or cancelled, or **30 days** | rows naming the person as invitee **or** inviter are **deleted** |
| `pod_suggestion_dismissals` | DB | **180 days** | the person's rows are **deleted**. Other people's rows hold only a one-way hash of a set that included them, which names nobody and goes in 180 days |
| Suggestions | computed on read | not stored | nothing to delete |
| Invite rate counters | memory | 1 h | dropped |

- **`accountdelete.Inventory`** lists every column of the four tables. `Delete`, in its one transaction, passes or deletes owned pods, then removes the person's member, invite and dismissal rows. The audit line gains `pods_passed_on`, `pods_deleted`, `pod_members`, `pod_invites` and `pod_dismissals`.
- **`PRIVACY_SCHEMA_VERSION`** takes slice 2's migration number.
- **The privacy page** (`routes/Privacy.svelte`, numbers in `lib/privacy.ts`):
  - **Slice 2:** a new **"Pods"** entry under "What is stored, and why": "If you make or join a pod, we keep its name, who is in it and who owns it, and the invites you send or get (they expire after 30 days). The pod's members see your display name and avatar, that you are a member, and whether you share your stats with it. If you leave, are removed, or delete your account, your membership is gone at once. If you owned the pod, it passes to the member who joined first."
  - **Slice 4:** to that entry: "We may suggest a pod to people who have played 3 or more games together in the last 30 days. It is worked out from those games and not stored. It forms nothing unless each person accepts. If you choose Not now, we keep that choice for 180 days."
  - **Slice 5** (ADR 0217 slice 7), the pod sentence ADR 0217 §9 promises, under "Who it is shared with": "If you switch on Share with my pods, the members of every pod you are in can see your stats and your place on its leaderboard. People you have blocked, or who have blocked you, never see them through a pod. Joining a pod does not share anything by itself."
  - New numbers, pinned by `TestPrivacyNoticeNumbersMatchTheServer`: `podInviteDays` (30), `podDismissDays` (180), `podMemberCap` (12), `podsPerPerson` (10), `podSuggestGames` (3) and `podSuggestDays` (30).

## Calls made here

Each is overturnable in review:

- **One owner, who passes on automatically** to the member who joined first. No co-owners or moderators.
- **Any member can invite their own friends; only the owner manages the link.**
- **The link lasts 7 days, has no use count, and needs sign-in** even to see the pod's name.
- **Silent decline and silent leave,** as ADR 0213's Ignore and unfriend.
- **The suggestion signal:** the same 3 or 4 signed-in people at 3 or more ended games in 30 days. No pairs.
- **Dismissals last 180 days.**
- **The limits:** 12 members, 10 pods each, 20 pending invites, 10 invites an hour, 3 pods then 1 an hour.
- **A block between members hides each from the other** and removes neither. The pod's numbers are unchanged.
- **One sharing switch for all pods,** as ADR 0217 built `share_pods`.
- **No report kind for pod names**, two admin routes instead.
- **Pod games include history** from before the pod was made.

## Alternatives considered

- **No owner: every member can do everything.** Simpler on paper, but a rename or removal fight then has no answer. One owner who can hand over is less code than a vote.
- **Co-owners or pod moderators.** A pod is at most 12 friends. Revisit if pods grow.
- **Friends only, no link.** Ruled out by owner answer 5. A link is also the only way in before ADR 0213's friends land for someone.
- **A link with no expiry,** like the stats share link. A share link shows one person's own stats; a pod link lets strangers *into* a group. Seven days, rotation and the size cap keep a posted link from becoming a door that stays open.
- **Removing one of two members who block each other.** It tells the blocked person, and it makes the pod's makeup depend on who blocks first. Hiding them from each other matches what a block does everywhere else (ADR 0213 §3).
- **Per-viewer pod numbers** (dropping a hidden member's games from "pod games" for the viewer). Each member would see a different record for the same group, and a missing count would hint at the block. Hiding the row is enough.
- **A per-pod sharing switch** (`pod_members.sharing`). Finer, but ADR 0217 shipped one column and the owner named it as the switch. A per-pod column is an additive change later if people ask.
- **Storing suggestions** in a table written nightly. It stores more about who plays with whom, and a read-time query over the caller's last 30 days of games is cheap at this size.
- **Pods as a Discord role or channel.** It would tie pods to the C2 server's moderators and to people outside the group, and it needs the bot to manage roles.
- **A new report kind for pod names now.** The audience is small and chose to be there, and the filter catches the listed words. It is an additive change if it is ever needed.

## Risks

- **A leaderboard can sour a group.** Mitigations: sharing is opt-in per person, a member can switch it off or leave at any time, rates need 5 games, and every row shows its count.
- **A bad pod name.** The filter is a slur list, not a judgement. The audience is small, any member can leave, and an admin can rename or delete a pod by id. If it happens more than rarely, add the report kind.
- **A posted invite link lets strangers in.** It needs sign-in, it expires in 7 days, the pod caps at 12, the owner can rotate it and remove anyone, and Remove offers a new link by default.
- **A block may still be guessed.** A member who vanishes from your list can be inferred. That is true of friends and Played with too (ADR 0213), and it is the cost of a block that tells nobody.
- **Suggestions can feel watchful.** They use only games the people shared, which each of them already sees in Played with. They store nothing until someone acts, and Not now hides one for 180 days. If they still feel wrong, slice 4 can be left out on its own.
- **Pods wait on ADR 0213 slice 3.** Friends and blocks are designed but not built. Slice 2 below needs them. See the owner question under Slices.

## Slices

Each slice is one mergeable PR with its tests, against `develop`.

| # | Slice | Schema | Server | Client | Privacy | Tier | Needs |
|---|---|---|---|---|---|---|---|
| 1 | **This ADR**, and its notes in ADR 0217 | — | — | — | — | — | — |
| 2 | **Pods core** | §9's migration | every route in §10 but suggestions and stats; the roles, the link, invites, leaving and removal; §6's block checks; the limits; the sweeper; the admin routes; docs/lobby.md. Tests: each role's allowed and refused actions, the owner hand-over on leave and on delete, the dead-link answer for every case, a block before and after joining | — | `Inventory` and `Delete`, the "Pods" entry, the numbers, the schema bump | 3 | ADR 0213 slice 3 |
| 3 | **Pod pages** | — | — | `#/pods`, `#/pods/<id>`, `#/pods/join/<token>`, the friend picker, the nav link, labels and e2e, screenshots | — | 2 | 2 |
| 4 | **Suggestions** | — (the dismissal table is in slice 2) | the read-time query in `GET /me/pods`, `POST /pods` with `suggestion`, the dismiss route. Tests: the signal on hand-built games, each exclusion | the Suggested section | the suggestion sentence and its numbers | 2 | 2, 3 |
| 5 | **Pod stats and the leaderboard** = ADR 0217 slice 7 | none | `GET /pods/{id}/stats`, `GET /pods/{id}/leaderboard`, §6's filter | `#/pods/<id>/stats`, the sharing switch | the pod sentence | 3 | 2, 3, and ADR 0217 slice 6 |

Slice 4 can wait or be dropped without touching the others.

**Owner answer (Ian, 2026-11-05):** build ADR 0213 slice 3 (friends, `user_blocks`, `social_prefs`) first. Slice 2 here waits for it, so pods ship with friend invites and block checks from the start, as this ADR assumes. ADR 0213 slice 2 (Played with) comes before its slice 3.

## Consequences

- C2 gains its first group concept, kept to a name, an owner and a member list.
- ADR 0217 slice 7 can be built: its routes have a membership table to check.
- No engine, wire or snapshot change.
- Friends (ADR 0213) become the main way into a pod, which is one more reason to build its slice 3.

## Back-out

| To undo | Do |
|---|---|
| Slice 5 | remove the two stats routes and the page. Sharing members' data is unchanged |
| Slice 4 | remove the suggestion query and route. Dismissal rows go with the next sweep, or by hand |
| Slice 3 | remove the pages. The server keeps working |
| Slice 2 | remove the routes. The four tables can be dropped by hand after rolling the binary back (an older binary refuses the newer schema, `ErrSchemaTooNew`; see docs/environments.md). `stats_prefs.share_pods` stays and does nothing |

### Slice 5 as built

Pod stats and the leaderboard (this ADR's slice 5 = ADR 0217 slice 7), 2026-11-09, #4880:

- **Routes** `GET /pods/{id}/stats` and `GET /pods/{id}/leaderboard` (`lobby/pod_stats.go`), `?range=all|90d|30d`. The block filter is §6 as written: a hidden member is out of the member list, the board and head-to-head, and the pod's own numbers are counted over every sharing member. [docs/lobby.md](../lobby.md#get-podsidstats-and-get-podsidleaderboard) has the bodies.
- **The switch** is `PUT /me/stats/prefs {"share_pods": …}` (`gamestats.PrefsStore.SetSharePods`), shown in Settings → Stats and on the pod's page by one component, `PodShareSwitch`. Migration 0030 already had the column, so there is **no migration and no `PRIVACY_SCHEMA_VERSION` bump**; the privacy page gains the pod sentence under "Who it is shared with" and "Stats settings" now names the switch.
- **Viewing needs no switch of the viewer's own.** The switch decides what a member shows to others, not what they may look at; the page nudges a viewer who is not sharing that they have no row. Overturn if the group should be reciprocal.
- **No makeup filter** on pod stats: §7 names only the range filter.
- **Page** `#/pods/<id>/stats` reuses `ChartFigure`, `BarChart`, `StatTile` and `StatsFilters` (now with `showMakeup={false}`); the Pod page gains a **Pod stats** link.

## Owner decisions

Ian. **Decided.**

1. **Forming a pod.** From friends (ADR 0213) and/or by an invite link. C2 may also suggest a pod to players who often play together. A suggestion forms nothing on its own: each member accepts (ADR 0217 owner answer 5).
2. **Consent.** Joining a pod is not consent to share stats. `stats_prefs.share_pods` is the switch (ADR 0217 §6). A member who is not sharing shows "not sharing" with no numbers.
3. **Who can be a member.** Not guests, bots or agents. Deleted accounts leave. ADR 0213 blocks apply.
4. **The leaderboard.** Pod games only, with no "All games" scope (this supersedes ADR 0217 §6's earlier text). Ranked by win rate, then average place, with a minimum of 5 counted games, and each row shows its count (ADR 0217 owner answer 7).
5. **The stats routes and page** are ADR 0217's: `GET /pods/{id}/stats`, `GET /pods/{id}/leaderboard`, members only, and `#/pods/<id>/stats`.
6. **Order (2026-11-05).** Friends and blocks (ADR 0213 slices 2 and 3) are built before this ADR's slice 2; pods are not shipped link-only first. The calls under [Calls made here](#calls-made-here), including the limits in §5 and blocks hiding two members from each other both ways (§6), stand as written.

## Sprint

Outside the numbered sprints' scope. The active sprint is S80 — Table effects: a cue for every common moment, and this is part of the owner's stats epic (Ian, #4880), as ADR 0217 and its slices were. It touches no engine rules.
