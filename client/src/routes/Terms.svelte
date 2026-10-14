<script lang="ts">
  // Terms — the terms of service (#3358): who runs the site, who may use
  // it, how to behave at a table, what happens to what you put on it,
  // copyright notices, the fan-content position, the source offer, and
  // the usual limits.
  //
  // Public on purpose (lib/signedInHome.ts), like the privacy notice: a
  // stranger has to be able to read it before signing in. The operator
  // and contact come from lib/legalNotice.ts, which the privacy notice
  // shares, and the other fixed values from lib/terms.ts. Statements about
  // what the site does (guests, reports, blocks, the agent seat, GitHub
  // issues) must stay true of the code, as on the privacy page.
  import Page from "../lib/components/Page.svelte";
  import LegalNotice from "../lib/components/LegalNotice.svelte";
  import {
    LEGAL_CONTACT_EMAIL,
    LEGAL_CONTACT_URL,
    OPERATOR_NAME,
    OPERATOR_STATE,
  } from "../lib/legalNotice";
  import {
    COPYRIGHT_AGENT_PHONE,
    COPYRIGHT_AGENT_POSTAL,
    LIABILITY_CAP_USD,
    MIN_AGE,
    TERMS_UPDATED,
  } from "../lib/terms";
  import { PRIVACY_HASH } from "../lib/privacy";
  import { REPO_URL } from "../lib/about";

  const LICENSE_URL = `${REPO_URL}/blob/main/LICENSE`;
</script>

<Page
  eyebrow="Terms"
  title="Terms of service"
  lede="These are the rules for using cmd_and_ctrl: who runs it, who may use it, how to behave at a table, and what happens to what you put on it. By using the site you agree to them. If you don't agree, please don't use it."
  width="prose"
>
  {#snippet meta()}
    <p class="meta">
      Last updated {TERMS_UPDATED}. Run by {OPERATOR_NAME}, a {OPERATOR_STATE} limited liability company.
      Contact: <a href={LEGAL_CONTACT_URL}>{LEGAL_CONTACT_EMAIL}</a>.
    </p>
  {/snippet}

  <div class="stack">
    <section aria-labelledby="who-title">
      <h2 id="who-title">Who you are agreeing with</h2>
      <p>
        cmd_and_ctrl (CMD &amp; CTRL, or C2) is run by {OPERATOR_NAME} (“we”, “us”). These terms are an
        agreement between you and {OPERATOR_NAME}. They cover the website, its game tables and the
        C2 Discord bot. What we store about you is described separately in the
        <a href={PRIVACY_HASH}>privacy notice</a>.
      </p>
    </section>

    <section aria-labelledby="who-may-title">
      <h2 id="who-may-title">Who may use it</h2>
      <p>
        You must be at least {MIN_AGE} years old, which is also Discord’s minimum. If you are under 18,
        you need a parent or guardian’s permission, and they agree to these terms on your behalf.
      </p>
      <p>
        An account comes from signing in with Discord, so Discord’s own terms apply to that account
        too. When sign-up is limited to members of the C2 Discord server who hold its account role,
        holding that role is a condition of having an account, and the server’s moderators decide
        who gets it. Without an account you can still read the public pages, check a deck, play a
        practice game, and sit at a table you were invited to.
      </p>
    </section>

    <section aria-labelledby="account-title">
      <h2 id="account-title">Your account</h2>
      <p>
        One account is for one person. You are responsible for what happens under your account and
        on any seat or spectator link you hold. Keep your Discord account secure, and don’t share
        invite or reclaim links with people you don’t want at the table.
      </p>
    </section>

    <section aria-labelledby="conduct-title">
      <h2 id="conduct-title">How to behave</h2>
      <p>When you use the site, don’t:</p>
      <ul>
        <li>
          harass, threaten or abuse other players, or post hateful, sexual or illegal content, in
          chat, names, deck names or images;
        </li>
        <li>pretend to be someone else, including another player, a moderator or the operator;</li>
        <li>
          use a bug on purpose to win or to disrupt a game. Report it instead, with the in-app
          button;
        </li>
        <li>
          play with software you control and pass it off as a person. The site’s own bots and the
          agent seat are allowed and are shown as such. Don’t hide that a seat is automated;
        </li>
        <li>
          scrape, flood or overload the site, try to get round its limits, or probe it for security
          holes (if you find one, tell us at the contact above);
        </li>
        <li>advertise, spam, or use the site to sell anything.</li>
      </ul>
      <p>
        A table’s host can remove and block people from their table, and any player can report
        another. Moderators and the operator may act on a report, which can mean a warning, losing
        the Discord account role, or losing access to the site.
      </p>
    </section>

    <section aria-labelledby="content-title">
      <h2 id="content-title">What you put on the site</h2>
      <p>
        You keep whatever rights you have in what you put here: your seat name, chat, deck lists,
        playmat images, reports, bug reports and deck requests. You give {OPERATOR_NAME} a worldwide,
        non-exclusive, royalty-free licence to store, copy, show and send that content as needed to run
        the site. For example, we show your playmat and your deck’s name to the other players at your
        table, keep the record and replay of games you played in, and back them up. The licence ends when
        the content is deleted, except for copies already in backups (which age out, as the privacy notice
        explains) and in other players’ game records.
      </p>
      <p>
        Bug reports, ideas, questions and deck requests are filed as issues in the project’s public
        GitHub repository. By sending one you agree that it is published there, with your display
        name.
      </p>
      <p>
        Only upload or link things you have the right to use. We may remove anything that breaks
        these terms or that someone tells us infringes their rights.
      </p>
    </section>

    <section aria-labelledby="copyright-title">
      <h2 id="copyright-title">Copyright notices</h2>
      <p>
        If you believe something on the site infringes your copyright, send a notice to our
        copyright agent at <a href={LEGAL_CONTACT_URL}>{LEGAL_CONTACT_EMAIL}</a> with:
      </p>
      <ol>
        <li>your physical or electronic signature;</li>
        <li>the work you say is infringed;</li>
        <li>
          the material you say infringes it, and where on the site it is (a link or a description we
          can find);
        </li>
        <li>your name, address, phone number and email address;</li>
        <li>
          a statement that you believe in good faith the use is not authorised by the owner, its
          agent or the law; and
        </li>
        <li>
          a statement that the notice is accurate and, under penalty of perjury, that you are the
          owner or are authorised to act for them.
        </li>
      </ol>
      <p>
        If your content was removed and you believe that was a mistake, you can send a
        counter-notice to the same address. It must include your signature, what was removed and
        where it was, a statement under penalty of perjury that you believe in good faith it was
        removed by mistake or misidentification, your name, address and phone number, and your
        consent to the jurisdiction of the federal court for your address (or, outside the United
        States, any judicial district where
        {OPERATOR_NAME} may be found) and to accept service from the person who sent the notice.
      </p>
      <p>We end the access of people who repeatedly infringe.</p>
      {#if COPYRIGHT_AGENT_POSTAL && COPYRIGHT_AGENT_PHONE}
        <p class="agent">
          Copyright agent: {OPERATOR_NAME}, {COPYRIGHT_AGENT_POSTAL}; phone {COPYRIGHT_AGENT_PHONE};
          email {LEGAL_CONTACT_EMAIL}.
        </p>
      {/if}
    </section>

    <section aria-labelledby="magic-title">
      <h2 id="magic-title">Magic: The Gathering and card data</h2>
      <div class="notice"><LegalNotice /></div>
      <p>
        These terms give you no rights in Wizards of the Coast’s material or Scryfall’s data. They
        belong to their owners. The rules engine is unofficial and can be wrong, and the site is not
        for sanctioned or prize play.
      </p>
    </section>

    <section aria-labelledby="source-title">
      <h2 id="source-title">The source code</h2>
      <p>
        cmd_and_ctrl is free software. Its source code is available at
        <a href={REPO_URL} target="_blank" rel="noopener noreferrer">GitHub</a> under the
        <a href={LICENSE_URL} target="_blank" rel="noopener noreferrer"
          >GNU Affero General Public License, version 3 or later</a
        >. Your rights to the code come from that licence, not from these terms. The licence covers
        the project’s own code. It does not cover Magic: The Gathering material or card data and
        images.
      </p>
    </section>

    <section aria-labelledby="bots-title">
      <h2 id="bots-title">Bots and AI players</h2>
      <p>
        Bot seats are played by software. Some may use a language model, which receives a
        description of the game as that seat sees it, as the privacy notice explains. An agent seat
        is an AI agent that someone has connected to the table, and is shown as one. Bots and agents
        can play badly or make mistakes, and we are not responsible for how they play.
      </p>
    </section>

    <section aria-labelledby="warranty-title">
      <h2 id="warranty-title">No guarantees</h2>
      <p>
        The site is a hobby project provided “as is” and “as available”, without warranties of any
        kind, express or implied, including fitness for a particular purpose and non-infringement.
        It may be down, slow or wrong. Games, decks and settings can be lost. Features can change or
        stop at any time.
      </p>
    </section>

    <section aria-labelledby="limits-title">
      <h2 id="limits-title">Limits on our liability</h2>
      <p>
        As far as the law allows, {OPERATOR_NAME} is not liable for indirect, incidental, special, consequential
        or punitive damages, or for lost data, profits or goodwill, arising from your use of the site.
        Our total liability for any claim about the site is limited to US${LIABILITY_CAP_USD}. Some
        places don’t allow these limits, and where they don’t, they apply only as far as allowed.
      </p>
      <p>
        If someone makes a claim against us because of content you posted or because you broke these
        terms, you agree to cover our reasonable costs of dealing with it, as far as the law allows.
      </p>
    </section>

    <section aria-labelledby="ending-title">
      <h2 id="ending-title">Ending your access</h2>
      <p>
        You can stop using the site at any time, and delete your account as the
        <a href={PRIVACY_HASH}>privacy notice</a> describes. We may suspend or end anyone’s access, remove
        content, or close the site, for any reason, especially for breaking these terms. The sections
        on your content’s licence, copyright, no guarantees, limits and the law that applies continue
        after your access ends.
      </p>
    </section>

    <section aria-labelledby="changes-title">
      <h2 id="changes-title">Changes to these terms</h2>
      <p>
        We may change these terms. The date at the top says when they last changed, and we will
        announce important changes on the site or the C2 Discord server. If you keep using the site
        after a change, you accept the new terms.
      </p>
    </section>

    <section aria-labelledby="law-title">
      <h2 id="law-title">The law that applies</h2>
      <p>
        These terms are governed by the laws of the State of {OPERATOR_STATE} and of the United States,
        without regard to conflict-of-law rules. Disputes go to the state or federal courts located in
        {OPERATOR_STATE}. If you live somewhere whose law gives you consumer rights that a contract
        can’t take away, nothing here takes them away. If any part of these terms can’t be enforced,
        the rest still applies.
      </p>
    </section>

    <p class="dim">
      Questions about these terms: <a href={LEGAL_CONTACT_URL}>{LEGAL_CONTACT_EMAIL}</a>.
    </p>
  </div>
</Page>

<style>
  .stack {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 24px;
  }
  /* The shared section head (the display face), as Privacy's and About's. */
  h2 {
    margin: 0 0 8px;
    font-family: var(--font-display);
    font-size: 19px;
    font-weight: 800;
    letter-spacing: -0.005em;
    text-transform: none;
    color: var(--fg);
  }
  p,
  li {
    max-width: 72ch;
    color: var(--fg-muted);
    font-size: 14px;
    line-height: 1.6;
  }
  p {
    margin: 0 0 8px;
  }
  ul,
  ol {
    margin: 0 0 8px;
    padding-left: 1.4em;
  }
  li + li {
    margin-top: 4px;
  }
  a {
    color: var(--accent);
  }
  .meta {
    margin: 8px 0 0;
    font-size: 14px;
  }
  .notice {
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    padding: 12px 14px;
    margin-bottom: 8px;
    font-size: 13px;
  }
  .notice :global(p) {
    font-size: 13px;
  }
  .dim {
    color: var(--fg-dim);
  }
</style>
