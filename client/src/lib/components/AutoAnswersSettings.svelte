<script lang="ts">
  // AutoAnswersSettings — Settings → Gameplay → Automatic answers (ADR
  // 0127 §6): one row per standing answer, with the card, the question,
  // Ask / Always / Never, and Forget. Choosing Ask removes the rule. The
  // rules are added from a prompt ("Remember this answer"); the server
  // answers with them, also while this tab is closed.
  import { L } from "../labels";
  import { settings, updateSettings, type AutoAnswerRule } from "../settings";
  import { withoutRule } from "../autoAnswerPref";

  const rules = $derived($settings.gameplay.autoAnswers);

  function setAnswer(rule: AutoAnswerRule, value: string): void {
    const list = $settings.gameplay.autoAnswers;
    if (value === "ask") {
      updateSettings("gameplay", "autoAnswers", withoutRule(list, rule.key));
      return;
    }
    if (value !== "always" && value !== "never") return;
    const answer: AutoAnswerRule["answer"] = value;
    updateSettings(
      "gameplay",
      "autoAnswers",
      list.map((r): AutoAnswerRule => (r.key === rule.key ? { ...r, answer } : r)),
    );
  }

  function forget(rule: AutoAnswerRule): void {
    updateSettings(
      "gameplay",
      "autoAnswers",
      withoutRule($settings.gameplay.autoAnswers, rule.key),
    );
  }
</script>

<section class="auto-answers" aria-label={L.automaticAnswers}>
  <h4>Saved answers</h4>
  {#if rules.length === 0}
    <p class="help">
      None yet. Tick {L.rememberThisAnswer} on a prompt, and the next button you press becomes the standing
      answer for that card's question.
    </p>
  {:else}
    <p class="help">
      The game answers these for you, also while this tab is closed, and writes "(automatic)" in the
      log for everyone. "Always pay" pays only when your untapped mana covers it.
    </p>
    <ul>
      {#each rules as rule (rule.key)}
        <li>
          <span class="what">
            <span class="card">{rule.card || "A card"}</span>
            {#if rule.prompt}<span class="prompt">{rule.prompt}</span>{/if}
          </span>
          <select
            aria-label={`Answer for ${rule.prompt || rule.card || "this prompt"}`}
            value={rule.answer}
            onchange={(e) => setAnswer(rule, e.currentTarget.value)}
          >
            <option value="ask">Ask</option>
            <option value="always">Always</option>
            <option value="never">Never</option>
          </select>
          <button type="button" class="forget" onclick={() => forget(rule)}>
            {L.forgetAutoAnswer}
          </button>
        </li>
      {/each}
    </ul>
  {/if}
</section>

<style>
  .auto-answers h4 {
    margin: 0.75rem 0 0.25rem;
  }
  ul {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
  }
  li {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    flex-wrap: wrap;
  }
  .what {
    display: flex;
    flex-direction: column;
    flex: 1 1 12rem;
    min-width: 0;
  }
  .card {
    font-weight: 600;
  }
  .prompt {
    opacity: 0.8;
    font-size: 0.9em;
    overflow-wrap: anywhere;
  }
</style>
