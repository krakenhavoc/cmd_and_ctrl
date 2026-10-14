<script lang="ts">
  // LegalNotice — the full fan-content notice (#2191, #2982), one copy
  // for both places it is shown: the site footer's "Legal" disclosure and
  // the table's ⋯ menu (the footer is hidden on the game route, so the
  // table needs its own way to the same words). Plain paragraphs and a
  // link; the caller supplies the frame and the type size.
  import { FAN_CONTENT_POLICY_URL } from "../legalNotice";
  import { PRIVACY_HASH } from "../privacy";
  import { TERMS_HASH } from "../terms";

  // newTab: the table's copy opens the privacy notice in a new tab, so
  // reading it never takes a player away from a game in progress.
  const { newTab = false }: { newTab?: boolean } = $props();
</script>

<p>
  cmd_and_ctrl is unofficial Fan Content permitted under the
  <a href={FAN_CONTENT_POLICY_URL} target="_blank" rel="noopener noreferrer">Fan Content Policy</a>.
  Not approved or endorsed by Wizards of the Coast. Portions of the materials used are property of
  Wizards of the Coast. ©Wizards of the Coast LLC. Magic: The Gathering is a trademark of Wizards of
  the Coast LLC. Card data and card images are provided by
  <a href="https://scryfall.com" target="_blank" rel="noopener noreferrer">Scryfall</a>, which is
  not affiliated with this site.
</p>

<p>
  <a
    href={PRIVACY_HASH}
    target={newTab ? "_blank" : undefined}
    rel={newTab ? "noopener noreferrer" : undefined}>Privacy and your data</a
  >: what this site stores, for how long, and how to delete your account.
  <a
    href={TERMS_HASH}
    target={newTab ? "_blank" : undefined}
    rel={newTab ? "noopener noreferrer" : undefined}>Terms of service</a
  >: who runs the site and the rules for using it.
</p>
