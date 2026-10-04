<script lang="ts">
  // The AI-agent chip (ADR 0122 §7): a seat played by an outside model,
  // shown to everyone. Deliberately unlike the bot chip. A bot is the
  // server's own seat (dashed, seat-tinted, robot glyph); an agent is a
  // separate program, so it is violet, solid, with a spark glyph and the
  // client's name. The chip is a role="img" with an accessible name
  // rather than bare text, because its visible text ("AI · claude-code",
  // "thinking…") is an abbreviation of what it states.
  import Icon from "./Icon.svelte";
  import { agentChipLabel, agentChipText, agentChipTitle, type AgentFields } from "../agentSeat";

  interface Props {
    seat: AgentFields;
    /** The seat holds priority: the chip reads "thinking…". */
    thinking?: boolean;
  }
  const { seat, thinking = false }: Props = $props();
</script>

<span
  class="agent-chip"
  class:thinking
  role="img"
  aria-label={agentChipLabel(seat, thinking)}
  title={agentChipTitle(seat)}
>
  <Icon name="spark" size={9} />
  <span aria-hidden="true">{agentChipText(seat, thinking)}</span>
</span>

<style>
  .agent-chip {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    max-width: 100%;
    box-sizing: border-box;
    padding: 2px 7px;
    border-radius: 4px;
    border: 1px solid var(--agent-border, rgba(176, 138, 255, 0.55));
    background: var(--agent-soft, rgba(176, 138, 255, 0.16));
    color: var(--agent, #c3a6ff);
    font-family: var(--font-mono);
    font-size: 9px;
    font-weight: 700;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    white-space: nowrap;
    vertical-align: middle;
  }
  .agent-chip span {
    overflow: hidden;
    text-overflow: ellipsis;
  }
</style>
