<script lang="ts">
  // Privacy — what the site stores about a person, why, for how long, who
  // else sees it, and how to delete it (#2983, ADR 0149).
  //
  // Public on purpose (lib/signedInHome.ts): someone deciding whether to
  // sign in has to be able to read it first. It holds no account data.
  //
  // EVERY STATEMENT HERE IS DERIVED FROM THE CODE. The numbers come from
  // lib/privacy.ts, which the Go test server/internal/lobby/
  // privacy_facts_test.go checks against the server constants, and the
  // table of what is stored follows accountdelete.Inventory, which a Go
  // test checks against the live database schema. When you add anything
  // that stores something about a person, update this page in the same
  // change (ADR 0149 §6). The source of each row is cited in its
  // `source` field below; the test does not read it, a reviewer does.
  import Page from "../lib/components/Page.svelte";
  import DeleteAccount from "../lib/components/DeleteAccount.svelte";
  import { CONTACT_TEXT, CONTACT_URL, PRIVACY_FACTS as F } from "../lib/privacy";
  import { OPERATOR_NAME, OPERATOR_STATE } from "../lib/legalNotice";
  import { TERMS_HASH } from "../lib/terms";
  import { SHOP_DISCLOSURE } from "../lib/shopLinks";

  interface Row {
    what: string;
    details: string;
    why: string;
    howLong: string;
    /** Where in the code this is true. Not rendered; for the reviewer. */
    source: string;
  }

  const stored: Row[] = [
    {
      what: "Your Discord sign-in",
      details:
        "Your Discord user id, display name and avatar, and the time you last signed in. We ask Discord for the “identify” permission: no email address and no server list. When sign-up needs the account role in the C2 Discord server (Elder Dragon Highlander), we also ask for “guilds.members.read”, which Discord describes as reading your member info (nickname, avatar, roles) in servers you are in. The server uses it only to read your own membership of the C2 server, and from it only whether you hold that role: it keeps the answer (admitted, not a member, or no role) and when it was given, and nothing else from that record. It asks again each time your sign-in is renewed. If Discord turns you away at sign-in, nothing is kept about you. When the server has an encryption key set, Discord’s refresh token is kept, encrypted, and is what lets the server ask again at renewal; otherwise it is thrown away.",
      why: "To make you the same person on every device, put your name and picture on your seat, keep your decks and games with you, and keep accounts to members of the C2 Discord.",
      howLong: "Until you delete your account.",
      source:
        "migrations 0003, 0019 users.access_checked_at, users.access_result; discord/oauth.go Scopes; discord/gate.go CheckMember; lobby/access_gate.go",
    },
    {
      what: "Your session",
      details:
        "A signed token, kept in a cookie called cmdctrl_session and in this browser’s local storage. The server keeps no list of sessions; it keeps only the time you last chose “sign out everywhere”.",
      why: "So the server knows who is asking.",
      howLong: `${F.signInDays} days after a Discord sign-in. ${F.guestSessionHours} hours for a guest seat or a spectator link.`,
      source: "auth/http.go SessionCookie; lobby/http.go identityTTL, sessionTTL",
    },
    {
      what: "Your decks",
      details:
        "Each saved deck’s name, the list you pasted or the link you imported it from, its commanders, and its estimated power level once an admin rates it (shown to the table you seat it at).",
      why: "So you can seat a deck without pasting it again.",
      howLong: "Until you delete the deck or your account.",
      source: "migrations 0005, 0008, 0027",
    },
    {
      what: "Your settings and last table",
      details:
        "Your synced settings (a small document, at most 32 KiB), and the last table you started: its table settings, its bot seats and the other signed-in players who sat at it.",
      why: "To bring your preferences to a new device and set up your next table the same way.",
      howLong: "Until you delete your account.",
      source: "migration 0008 user_settings, table_setups, users.last_deck",
    },
    {
      what: "Your Discord notifications",
      details:
        "Whether each of the three Discord direct-message notifications is on (your turn, table ready, table filled), all off until you turn one on, and the time Discord last refused a message to you, if it has. While you play, the server also keeps in memory whether your table’s browser tab is in front, when you last acted, and when it last sent you a message; none of that is written down.",
      why: "So the server only messages you when you asked it to, stops when Discord says it cannot reach you, and sends at most one message per table every ten minutes.",
      howLong:
        "Until you delete your account. What is kept in memory is gone when the server restarts.",
      source: "migration 0022 notification_prefs; notify/notifier.go; ws/presence.go",
    },
    {
      what: "Your own model key for AI opponents",
      details:
        "If you add one in Settings › AI opponents: the API key, stored encrypted, and its last four characters; the provider and model you chose (and the address, for another OpenAI-compatible provider); your games-a-day and calls-per-game limits; how many games used it today; when you added it and last used it; and the last error, as one of CMD & CTRL's own sentences. A bot seat you ran on it records that it was yours and the model it asked for. The key is never shown again, logged, or put in a replay.",
      why: "So the AI seats you add play on your account with the provider, used only for your own seats. Each request goes to the provider you chose, under its own terms and privacy policy, and carries the game the bot is looking at.",
      howLong:
        "Until you remove the key or delete your account. Deleting your account deletes the key; a past bot seat then points at the anonymous “Deleted player”.",
      source: "migration 0025 model_keys, seats.bot_key_user_id, seats.bot_key_model; modelkeys/",
    },
    {
      what: "Your playmats",
      details: `Up to ${F.playmatSlots} images you upload (or fetch once from a link you give), stored as picture files, and which one is active.`,
      why: "So the table can show your playmat to the other players.",
      howLong: "Until you remove them or delete your account.",
      source: "migrations 0011 to 0013; playmat/fit.go MaxSlots",
    },
    {
      what: "Games you played in",
      details:
        "The table’s name, who sat in which seat (the seat’s name, and the account it belongs to), the deck name, bot seats, and when it started and ended and how. For each game the server also keeps files: the saved state of the table, a replay of every move, and the last view of the board.",
      why: "So a table survives a restart, can be rejoined, and can be looked at and replayed after it ends.",
      howLong: `A game’s record and its replay are kept until a server admin deletes the table; there is no automatic limit. The saved state of an ended table is rebuilt for ${F.endedTableDays} days after it ends and removed after that. Deleting your account turns your seat into “Deleted player” in all of it.`,
      source:
        "migrations 0002, 0004, 0007, 0010; ws/persist.go EndedTableRetention; ws/room.go snapshotPath, replayPath",
    },
    {
      what: "Game statistics",
      details:
        "When a table ends, the server works out numbers about the game from its record: how many turns it took, and for each seat its result and place, when it was knocked out, its life gained and lost, the damage it dealt, the spells it cast and which cards by name, the cards it drew, its attacks, tokens and mulligans, and its commanders. They are stored by game and seat, with no names, and point at your seat through the game record above. The Stats page shows you the numbers for your own seats, and nobody else’s. Server admins can read the numbers for all games, with the display names of the people who sat at them, on an admin page.",
      why: "So you can see how your decks and games go, on your Stats page.",
      howLong:
        "As long as the game’s record is kept. Deleting the table deletes them. Deleting your account turns your seat into “Deleted player”, and these numbers then belong to that seat, as the rest of the game does.",
      source:
        "migration 0023 game_stats, seat_stats, seat_card_stats; gamestats/gamestats.go Compute",
    },
    {
      what: "Results post in Discord",
      details:
        "Only if the host of a table turns it on, and only on a server that has a results channel: when the game ends, the bot posts the table’s name, each seat’s name, its commander names and how it placed, the turn count and how long the game ran, in the CMD & CTRL Discord server’s results channel. Seat names are whatever each person set at the table, guests and AI agents included. Nobody is mentioned or pinged. A table’s page shows when this is on.",
      why: "So a finished game can be shown off to the community.",
      howLong:
        "The server keeps only the host’s on or off choice, with the table. The post itself is a Discord message that stays until a Discord moderator deletes it; deleting your account or the table does not remove a post already made.",
      source: "migration 0024 games.results; lobby/results_feed.go; discord/results.go",
    },
    {
      what: "Chat",
      details: `Table chat is not stored. The server holds the last ${F.chatRecentLines} lines of each table in memory, with the seat names they were written under, and only so that a player report can attach them. They are gone when the server restarts. A listed slur is replaced with asterisks before anyone sees the line.`,
      why: "So a report about what someone said has the words in it, and so slurs do not reach the table.",
      howLong:
        "In memory only, until the server restarts. A copy is kept only inside a report (see below).",
      source: "ws/chat_recent.go RecentChatLimit; moderation/filter.go",
    },
    {
      what: "Player reports",
      details: `When a player at your table reports another, the server keeps both parties’ names and account ids (a spectator can be reported too), the reporter’s one-line reason, the table’s name, and up to ${F.reportChatLines} recent chat lines (everyone’s, with the seat names they were written under). Only server admins can read it. On a server with a mod log, the bot also posts the table’s name, the reported player’s name and the reason (never who reported, the chat lines or any id) to a private channel only the CMD & CTRL Discord server’s moderators can read. Nobody is told who reported, and nothing happens to the reported player unless an admin decides it should. One person can file ${F.reportsPerWindow} reports in a burst, then one every ${F.reportEveryMinutes} minutes.`,
      why: "So the operator can see abuse at a public table and act on it.",
      howLong: `${F.reportDays} days after it is filed, or sooner if more than ${F.reportMax} are waiting (the oldest go first), or when an admin dismisses it. Deleting your account removes every report you filed or are the subject of; a report that only quotes one of your chat lines keeps the line until it expires. A mod log post is a Discord message that stays until a Discord moderator deletes it.`,
      source:
        "migration 0016 player_reports; lobby/reports.go ReportRetention, ReportMaxRows; lobby/modlog.go",
    },
    {
      what: "Tables that removed you",
      details:
        "When a table’s host (or an admin) removes someone, seated or watching, and chooses to block them, that table keeps what it needs to refuse them again: for a signed-in person, their account id, Discord id and the display name they had at the table; for a guest, their name folded to lower case without spacing or accents. Removing a spectator also keeps a random id for that one viewing session, which names no one. It is per table. The host and the creator of an unstarted table, and server admins, can see the list by that name (never a Discord id) and lift a block.",
      why: "So a removed person cannot sit straight back down with the same invite link, and so the host can undo a mistake.",
      howLong:
        "Until the table is deleted, until the host or an admin lifts the block, or until you delete your account, which removes every block entry that names your account along with the display name kept in it (a guest’s folded name names no account and goes with the table).",
      source: "migrations 0016 and 0017 table_blocks; lobby/blocks.go, lobby/spectators.go",
    },
    {
      what: "Deck requests",
      details: `When you ask for a deck’s missing cards, a row with your Discord id counts your requests (${F.deckRequestsPerWindow} every ${F.deckRequestWindowHours} hours). The request itself is a GitHub issue that names you by display name (see GitHub below). The server also remembers which cards each deck’s issue listed, with the deck’s link or a hash of its list but not your name, so an updated deck does not ask for the same card twice.`,
      why: "To limit how many requests one person can make, and to ask only for cards an issue does not already list.",
      howLong:
        "The count rows: until you delete your account. The list of cards an issue named stays with the deck’s issue, since it names no one. The GitHub issue: see below.",
      source:
        "migrations 0006 and 0028; lobby/deckcheck.go deckRequestLimit, “Requested by”; lobby/mydecks_update.go requestedSet",
    },
    {
      what: "Bug reports, ideas and questions",
      details: `What you type, your name and role, the table id and turn, and, for a bug or a question, the browser’s recent log. Up to ${F.bugImages} screenshots you attach, which can include a picture of the table as you see it, your own hand too, if you tick the box for it. A bug report or a question also keeps the table’s public game log, and a bug report a copy of its replay, both readable by server admins only.`,
      why: "So the problem can be understood and fixed.",
      howLong: `The screenshots and the saved copies are removed after ${F.bugArtifactDays} days, or sooner, oldest first, if the store passes ${F.bugArtifactMiB} MiB. They are not tied to your account. The GitHub issue stays (see below).`,
      source:
        "lobby/bugreport.go bugKinds; bugstore/store.go DefaultRetention, DefaultMaxStoreBytes",
    },
    {
      what: "Server logs and counts",
      details:
        "Server logs hold internal ids and the network address a connection came from. The counts the operator watches (how many games, how many people) carry no names.",
      why: "To find and fix faults and to see whether the site is healthy.",
      howLong: `Logs: ${F.logDays} days in the monitoring stack, which only the operator can open. The server host also keeps its own system journal under the host’s normal rotation, which this project does not set. Counts: a year.`,
      source: "ADR 0123 decisions 6 and 7; ws/hub.go (remote address in log lines)",
    },
    {
      what: "Backups",
      details:
        "When the operator has switched backups on: a nightly encrypted copy of the database and the game files, stored with Cloudflare.",
      why: "To recover from a lost disk.",
      howLong: `The last ${F.backupDaily} daily, ${F.backupWeekly} weekly and ${F.backupMonthly} monthly copies are kept. Deleting your account does not reach into past backups; they age out, the oldest in about ${F.backupMonthly} months.`,
      source: "scripts/backup-offsite.sh restic forget; docs/environments.md#backups",
    },
  ];

  interface Party {
    who: string;
    what: string;
  }

  const sharedWith: Party[] = [
    {
      who: "The other people at your table",
      what: "Your seat name and avatar, your playmat, the deck name, and everything that happens in the game. A spectator link shows the same. The players at a table can see the names of the spectators watching it (a signed-in spectator’s is their Discord display name), so they can report or ask the host to remove one.",
    },
    {
      who: "Discord",
      what: "Sign-in goes through Discord, which learns that you signed in to this site. After that the server, not your browser, fetches your avatar picture from Discord’s image servers and keeps a copy. When sign-up needs that role, the server also asks Discord whether you hold it, at sign-in and each time your sign-in is renewed. When a table’s host turns on a voice room, the server asks Discord to make a private voice channel in the C2 Discord server, named after the table, and lets in the seated players who are members of that server, by the Discord ids it already holds. To do that it asks Discord whether each of them is a member, and whether they are connected to that channel and muted, about every twenty seconds while the room is open; the seated players of that table see who is connected and who is muted, and nobody else does. The channel and its text chat are deleted after the game. The server’s owner and administrators can see every channel there, private ones included. When a table’s host posts it to #lfg, the bot posts the table’s name, the host’s Discord account and the seat count in that channel, where every member of the CMD & CTRL Discord server can see it; the post is deleted when the game starts or the host withdraws it. When you file a bug report, idea or question from outside a table, or from a table whose host posted it to #lfg, and the server has a bug forum, the bot also opens a thread there with what you wrote, links to your screenshots and a link to the GitHub issue, and mentions you if you signed in with Discord. Every member of the CMD & CTRL Discord server who can see the forum can read it. The bot closes the thread when the issue is closed; it stays until a Discord moderator deletes it. Reports from any other table are never posted there. When a player is reported, or an admin ends or deletes a table, removes a seat or a spectator, or signs an account out everywhere, the bot posts a line to a private mod log channel that only the server’s moderators can read: the table’s name, the name the person had at that table (or, for signing out, their display name), the reason for a report, and which admin acted; never who reported, chat lines or any id. It stays until a Discord moderator deletes it. When you ask for an account in the access-request channel, the bot replies with your request for a moderator to approve or decline; approving gives you the role that lets you sign in, and the request then says which moderator answered. If you turn on notifications, the server asks Discord’s bot to send you direct messages naming the table and its turn number, with a link back to it; Discord holds those messages under its own policy.",
    },
    {
      who: "GitHub",
      what: `Bug reports, ideas, questions and deck requests are filed as issues in this project’s public repository. They carry your display name, what you wrote, and the browser log. A screenshot is hosted by this server and shown through GitHub’s image proxy, so GitHub’s servers fetch it; the link stops working after ${F.bugArtifactDays} days. Anyone can read a public issue.`,
    },
    {
      who: "The model provider, when model bots are on",
      what: "Bot seats at the “assisted” and “strong” tiers send a text description of the game as that seat sees it (its own hand, the public cards, life totals, the stack, its deck list and the legal moves) and the names of the players at the table to a language-model API: Anthropic’s, or an OpenAI-compatible endpoint the operator runs. No Discord ids, avatars, emails, sign-in tokens or chat messages are sent. With no model configured, which is the default, nothing is sent and those bots play on rules and a heuristic.",
    },
    {
      who: "Cloudflare",
      what: "Traffic reaches the server through a Cloudflare tunnel, so Cloudflare handles every request in transit, including the session cookie. Backups, when on, are stored in Cloudflare R2, encrypted.",
    },
    {
      who: "Google Fonts",
      what: "Your browser loads this site’s fonts from Google Fonts, so Google sees your network address and browser.",
    },
    {
      who: "Scryfall, Moxfield and Archidekt",
      what: "Card data and pictures come from Scryfall through this server’s cache, and a deck imported by link is fetched by the server. They see the server, not you.",
    },
    {
      // ADR 0179: lib/shopLinks.ts (the link is built in the browser from
      // card names), components/ShopLinks.svelte (rel="noopener
      // noreferrer"), settings.ts display.shopLinks. No server route
      // sees the click.
      who: "Card shops, if you follow a shop link",
      what: `Where the site is set up for them, the Decks and Catalog pages show links to buy the cards at a shop such as TCGplayer or Card Kingdom. ${SHOP_DISCLOSURE} Following one takes you to the shop’s own site, sometimes through the shop’s partner network on the way, and from there the shop’s own privacy terms and cookies apply. The link holds card names and nothing about you, and it opens without telling the shop which page you came from. CMD & CTRL does not track the click: following a link sends nothing to this server and nothing is recorded. You can hide the links in Settings, under Display. That choice is kept with your settings: with your account if you are signed in, otherwise only in this browser.`,
    },
    {
      // #3558: routes/Funding.svelte. Plain links with rel="noopener
      // noreferrer"; no server route sees the click.
      who: "GitHub Sponsors, Ko-fi or Patreon, if you follow a support link",
      what: "The funding page may link to a place where you can choose to support CMD & CTRL. Following a link takes you to that site, under its own terms, privacy policy and cookies. The link holds nothing about you and opens without telling that site which page you came from. CMD & CTRL does not track the click: following a link sends nothing to this server and nothing is recorded.",
    },
  ];
</script>

<Page
  eyebrow="Privacy"
  title="What is stored about you, and how to delete it"
  lede="cmd_and_ctrl is a hobby Magic: The Gathering Commander table. This page says what it stores, why, for how long, who else sees it, and how to delete it. We do not sell data, show ads or run analytics. The only cookie is the session cookie. If this page and the code ever disagree, that is a bug."
  width="prose"
>
  {#snippet meta()}
    <!-- The operator and its contact (#3358) come from lib/legalNotice.ts,
         the same constants the terms of service use. -->
    <p class="contact" id="contact">
      Run by {OPERATOR_NAME}, a {OPERATOR_STATE} limited liability company. Contact:
      <a href={CONTACT_URL}>{CONTACT_TEXT}</a>. See also the
      <a href={TERMS_HASH}>terms of service</a>.
    </p>
  {/snippet}
  <div class="stack">
    <section aria-labelledby="stored-title">
      <h2 id="stored-title">What is stored, and why</h2>
      <ul class="rows">
        {#each stored as r (r.what)}
          <li class="row">
            <h3>{r.what}</h3>
            <dl>
              <dt>What</dt>
              <dd>{r.details}</dd>
              <dt>Why</dt>
              <dd>{r.why}</dd>
              <dt>How long</dt>
              <dd>{r.howLong}</dd>
            </dl>
          </li>
        {/each}
      </ul>
      <p class="help">
        A guest seat has no account. The name you typed is stored with the game like any other seat.
      </p>
    </section>

    <section aria-labelledby="shared-title">
      <h2 id="shared-title">Who it is shared with</h2>
      <ul class="rows">
        {#each sharedWith as p (p.who)}
          <li class="row">
            <h3>{p.who}</h3>
            <p>{p.what}</p>
          </li>
        {/each}
      </ul>
    </section>

    <section aria-labelledby="delete-info-title">
      <h2 id="delete-info-title">Deleting your account</h2>
      <p class="help">
        A signed-in Discord account can be deleted from the account menu, or here. It removes your
        decks, deck-request history, playmats, settings, last table setup, notification settings and
        Discord sign-in, signs you out everywhere, and turns your seat in every table that has ended
        into “Deleted player”, so the other players keep their games. It is refused while you hold a
        seat at a table that has not ended.
      </p>
      <p class="help">
        Some things are not ours to delete. Issues already filed on GitHub are public; ask through
        the contact above and the owner will edit or remove one. Backups roll off on the schedule
        above. If you sign in with Discord again afterwards you get a new, empty account.
      </p>
      <div class="del">
        <DeleteAccount />
      </div>
    </section>

    <p class="help dim">
      Contact: <a href={CONTACT_URL}>{CONTACT_TEXT}</a>. Questions about what is stored are welcome.
    </p>
  </div>
</Page>

<style>
  .stack {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 26px;
  }
  /* The shared section head (the display face), as About's and the roadmap's. */
  h2 {
    margin: 0 0 10px;
    font-family: var(--font-display);
    font-size: 19px;
    font-weight: 800;
    letter-spacing: -0.005em;
    text-transform: none;
    color: var(--fg);
  }
  h3 {
    margin: 0 0 6px;
    font-size: 14px;
    font-weight: 700;
    color: var(--fg);
  }
  .contact {
    margin: 8px 0 0;
    font-size: 14px;
  }
  .contact a,
  .help a {
    color: var(--accent);
  }
  .help {
    margin: 10px 0 0;
    max-width: 72ch;
    color: var(--fg-muted);
    font-size: 13px;
    line-height: 1.6;
  }
  .dim {
    color: var(--fg-dim);
  }
  .rows {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 10px;
  }
  .row {
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    padding: 12px 14px;
  }
  .row p {
    margin: 0;
    color: var(--fg-muted);
    font-size: 13px;
    line-height: 1.55;
  }
  dl {
    margin: 0;
    display: grid;
    grid-template-columns: 5.5rem minmax(0, 1fr);
    gap: 4px 12px;
    font-size: 13px;
    line-height: 1.55;
  }
  dt {
    color: var(--fg-dim);
    font-family: var(--font-mono);
    font-size: 10.5px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    padding-top: 2px;
  }
  dd {
    margin: 0;
    color: var(--fg-muted);
  }
  .del {
    margin-top: 14px;
    background: var(--surface);
    border: 1px solid var(--danger-line);
    border-radius: var(--radius-lg);
    padding: 14px 16px;
    max-width: 640px;
  }
  @media (max-width: 560px) {
    dl {
      grid-template-columns: minmax(0, 1fr);
    }
    dt {
      padding-top: 6px;
    }
  }
</style>
