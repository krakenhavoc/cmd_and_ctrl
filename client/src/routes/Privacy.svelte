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
  import SiteHeader from "../lib/components/SiteHeader.svelte";
  import DeleteAccount from "../lib/components/DeleteAccount.svelte";
  import { CONTACT_TEXT, CONTACT_URL, PRIVACY_FACTS as F } from "../lib/privacy";

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
        "Your Discord user id, display name and avatar, and the time you last signed in. We ask Discord for the “identify” permission only: no email address and no server list. When the server has an encryption key set, Discord’s refresh token is kept, encrypted; otherwise it is thrown away.",
      why: "To make you the same person on every device, put your name and picture on your seat, and keep your decks and games with you.",
      howLong: "Until you delete your account.",
      source: "migrations 0003; discord/oauth.go discordScopes",
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
        "Each saved deck’s name, the list you pasted or the link you imported it from, and its commanders.",
      why: "So you can seat a deck without pasting it again.",
      howLong: "Until you delete the deck or your account.",
      source: "migrations 0005, 0008",
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
      what: "Deck requests",
      details: `When you ask for a deck’s missing cards, a row with your Discord id counts your requests (${F.deckRequestsPerWindow} every ${F.deckRequestWindowHours} hours). The request itself is a GitHub issue that names you by display name (see GitHub below).`,
      why: "To limit how many requests one person can make.",
      howLong: "The rows: until you delete your account. The GitHub issue: see below.",
      source: "migration 0006; lobby/deckcheck.go deckRequestLimit, “Requested by”",
    },
    {
      what: "Bug reports, ideas and questions",
      details: `What you type, your name and role, the table id and turn, and, for a bug or a question, the browser’s recent log. Up to ${F.bugImages} screenshots you attach. A bug report or a question also keeps the table’s public game log, and a bug report a copy of its replay, both readable by server admins only.`,
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
      what: "Your seat name and avatar, your playmat, the deck name, and everything that happens in the game. A spectator link shows the same.",
    },
    {
      who: "Discord",
      what: "Sign-in goes through Discord, which learns that you signed in to this site. After that the server, not your browser, fetches your avatar picture from Discord’s image servers and keeps a copy.",
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
  ];
</script>

<SiteHeader />

<section class="entry">
  <div class="stack">
    <div class="head">
      <p class="eyebrow">Privacy</p>
      <h1>What is stored about you, and how to delete it</h1>
      <p class="lede">
        cmd_and_ctrl is a hobby Magic: The Gathering Commander table run by one person. This page
        says what it stores, why, for how long, who else sees it, and how to delete it. We do not
        sell data, show ads or run analytics. The only cookie is the session cookie. If this page
        and the code ever disagree, that is a bug.
      </p>
      <!-- TODO(owner): contact. No address has been chosen, so this points at
           the project's GitHub rather than inventing one. -->
      <p class="contact" id="contact">
        Contact: <a href={CONTACT_URL} target="_blank" rel="noopener noreferrer">{CONTACT_TEXT}</a>.
      </p>
    </div>

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
        decks, deck-request history, playmats, settings, last table setup and Discord sign-in, signs
        you out everywhere, and turns your seat in every table that has ended into “Deleted player”,
        so the other players keep their games. It is refused while you hold a seat at a table that
        has not ended.
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
      Contact: <a href={CONTACT_URL} target="_blank" rel="noopener noreferrer">{CONTACT_TEXT}</a>.
      Questions about what is stored are welcome.
    </p>
  </div>
</section>

<style>
  .entry {
    position: relative;
    min-height: calc(100vh - 3rem - 68px);
    display: flex;
    justify-content: center;
  }
  .stack {
    width: min(860px, 100%);
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 26px;
    margin: clamp(8px, 2vh, 24px) 0 48px;
  }
  .eyebrow {
    margin: 0 0 6px;
    font-family: var(--font-mono);
    font-size: 10.5px;
    letter-spacing: 0.16em;
    text-transform: uppercase;
    color: var(--accent-strong);
  }
  h1 {
    margin: 0;
    font-family: var(--font-display);
    font-weight: 800;
    letter-spacing: -0.01em;
  }
  h2 {
    margin: 0 0 10px;
    font-size: 18px;
    letter-spacing: -0.01em;
    text-transform: none;
    color: var(--accent);
  }
  h3 {
    margin: 0 0 6px;
    font-size: 14px;
    font-weight: 700;
    color: var(--fg);
  }
  .lede {
    margin: 10px 0 0;
    max-width: 68ch;
    color: var(--fg-muted);
    font-size: 14px;
    line-height: 1.6;
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
