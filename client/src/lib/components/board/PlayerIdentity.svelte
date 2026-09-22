<script lang="ts">
  // PlayerIdentity is the circular, identity-first rebuild of PlayerHeader.
  // It replaces the horizontal pill in the panel with an avatar disc as
  // the visual anchor: name above, life over the circle, markers +
  // ±-controls arranged around it. Same behaviors as the old header —
  // click-to-attack, click-to-cast-target, life/poison/energy steppers,
  // monarch/initiative toggles, damage floaters, mana pool pips.
  //
  // The avatar wrapper carries data-seat-id so CombatArrows can anchor
  // attack arrows on it (see CombatArrows.svelte:161).

  import type { ActionPayload, ActionType, PlayerView } from "../../protocol";
  import { seatColor } from "../../colors";
  import { floatUp, fadeOut } from "../../animations";
  import { play } from "../../sounds";
  import { avatarURL } from "../../api";
  import { scryfallImageURL } from "../../cardImage";
  import { targeting, isLegalPlayerTarget, isPicked } from "../../targeting";
  import { settings } from "../../settings";
  import { emptyLifeTracker, lifePopupView, trackLife, type LifePopupView } from "../../lifePopup";
  import { botDeckNames, botDeckLabel, ensureBotDeckNamesLoaded } from "../../botDeckNames";
  import { playerKeywordBadges } from "../../playerKeywordBadges";
  import ManaPoolPips from "./ManaPoolPips.svelte";
  import Icon from "../Icon.svelte";

  type ActionSender = (type: ActionType, params?: ActionPayload["params"], player?: string) => void;

  interface Props {
    seat: PlayerView;
    isSelf: boolean;
    isActive: boolean;
    hasPriority: boolean;
    attackTargetable: boolean;
    isMonarch: boolean;
    isInitiative: boolean;
    sendAction: ActionSender;
    onDeclareAttack?: (targetPlayerID: string) => void;
    onTargetPlayer?: (targetPlayerID: string) => void;
    // Scryfall id of the seat's commander (command zone or
    // battlefield), for the art-crop avatar when the seat has no
    // Discord avatar. Precedence: Discord → commander art → seat disc.
    commanderScryfallID?: string | null;
  }

  const {
    seat,
    isSelf,
    isActive,
    hasPriority,
    attackTargetable,
    isMonarch,
    isInitiative,
    sendAction,
    onDeclareAttack,
    onTargetPlayer,
    commanderScryfallID = null,
  }: Props = $props();

  const targetableByCast = $derived.by(() => {
    const t = $targeting;
    if (!t) return false;
    if (!isLegalPlayerTarget(t, seat.id)) return false;
    return !seat.eliminated;
  });
  // S20 sub-PR 5: already in a multi-target pick list.
  const pickedByCast = $derived.by(() => {
    const t = $targeting;
    return t !== null && isPicked(t, seat.id);
  });

  function handleAvatarClick(): void {
    if (targetableByCast) {
      onTargetPlayer?.(seat.id);
      return;
    }
    if (!attackTargetable) return;
    onDeclareAttack?.(seat.id);
  }

  const discordAvatar = $derived(avatarURL(seat.discord_id, seat.discord_avatar_hash));
  // The commander's art crop is always the FRONT face's — a
  // double-faced commander is identified by the side it is cast as,
  // and the seat header is an identity badge, not a board state.
  const commanderArt = $derived(
    commanderScryfallID ? scryfallImageURL(commanderScryfallID, "art_crop") : null,
  );
  // Discord avatar first, the commander's art crop when there is
  // none, the seat-colour disc when neither loads.
  let failedAvatarURL = $state<string | null>(null);
  const avatar = $derived.by(() => {
    if (discordAvatar && failedAvatarURL !== discordAvatar) return discordAvatar;
    if (commanderArt && failedAvatarURL !== commanderArt) return commanderArt;
    return null;
  });
  const displayLabel = $derived(seat.display_name ?? seat.name);

  // #1201: hexproof / protection-on-a-player badges (CR 702.11d, CR
  // 702.16i). See playerKeywordBadges.ts for the token → badge
  // mapping; this component only renders what it returns.
  const keywordBadges = $derived(playerKeywordBadges(seat.keywords));

  // --- bot seats (S31, ADR 0033) ---------------------------------
  //
  // A bot seat gets three marks: a BOT chip under the name, a robot
  // glyph in place of the avatar (a bot has no Discord portrait and
  // showing its commander's art would read as a human who picked
  // that commander), and a pulse while it is the seat we are waiting
  // on.
  const isBot = $derived(seat.is_bot === true);
  const botLabel = $derived(seat.bot_tier ? `bot · ${seat.bot_tier}` : "bot");
  // #688: PlayerView only carries the curated deck's ID (e.g.
  // "esper-control"); the display name comes from GET /bot/options,
  // resolved and cached client-side by botDeckNames.ts. An ID the
  // catalog doesn't recognize (a pasted custom decklist, or a deck
  // retired since the seat was made) falls back to the raw ID rather
  // than rendering nothing.
  $effect(() => {
    if (isBot) void ensureBotDeckNamesLoaded();
  });
  const botTitle = $derived(
    [
      seat.bot_tier ? `${seat.bot_tier} tier` : null,
      seat.bot_deck ? `deck: ${botDeckLabel(seat.bot_deck, $botDeckNames)}` : null,
    ]
      .filter(Boolean)
      .join(" — ") || "a bot plays this seat",
  );
  // "Thinking" is the honest signal we have without a new frame: the
  // bot holds priority, so the table is waiting on its runner. It
  // clears the moment priority moves.
  const botThinking = $derived(isBot && hasPriority && !seat.eliminated);
  // S11.5: the pulse is an animation, so it obeys the animations
  // toggle and the reduce-motion preference. With either off the chip
  // still says "thinking" — the information survives, the motion does
  // not.
  const animateThinking = $derived(
    $settings.animations.enabled && !$settings.accessibility.reduceMotion,
  );

  function changeLife(delta: number): void {
    sendAction("change_life", { delta }, seat.id);
  }
  function changePoison(delta: number): void {
    sendAction("add_player_counter", { name: "poison", delta }, seat.id);
  }
  function changeEnergy(delta: number): void {
    sendAction("add_player_counter", { name: "energy", delta }, seat.id);
  }
  function toggleMonarch(): void {
    sendAction("set_monarch", undefined, isMonarch ? "" : seat.id);
  }
  function toggleInitiative(): void {
    sendAction("set_initiative", undefined, isInitiative ? "" : seat.id);
  }

  // --- the life-change popup (#703) ----------------------------
  //
  // All the deciding lives in lifePopup.ts, over the seq the server
  // stamps on each life_history entry. This component only holds the
  // watermark and the hold timer. It deliberately does NOT diff the
  // log's length: a frame can carry several changes (double strike
  // lands both combat damage steps at once), and the log stops
  // growing at its server-side cap.
  let tracker = emptyLifeTracker();
  let popup = $state<{ key: number; view: LifePopupView } | null>(null);
  let popupTimer: ReturnType<typeof setTimeout> | null = null;
  $effect(() => {
    const frame = trackLife(tracker, seat.life_history);
    tracker = frame.tracker;
    if (!frame.popup) return;
    const view = lifePopupView(frame.popup);
    popup = { key: frame.popup.key, view };
    play(view.sound);
    if (popupTimer) clearTimeout(popupTimer);
    popupTimer = setTimeout(() => {
      popup = null;
      popupTimer = null;
    }, view.holdMs);
  });
  // Separate, dependency-free effect so the teardown runs on destroy
  // only — putting it on the effect above would clear the timer on
  // every frame.
  $effect(() => {
    return () => {
      if (popupTimer) clearTimeout(popupTimer);
      popupTimer = null;
    };
  });

  const interactive = $derived(attackTargetable || targetableByCast);
</script>

<div
  class="identity"
  class:self={isSelf}
  class:active={isActive}
  class:priority={hasPriority}
  class:targetable={attackTargetable}
  class:cast-targetable={targetableByCast}
  class:cast-picked={pickedByCast}
  class:eliminated={seat.eliminated}
  class:bot={isBot}
  class:thinking={botThinking && animateThinking}
  style:--seat-color={seatColor(seat.seat)}
>
  <span class="name" title={displayLabel}>{displayLabel}</span>
  <!-- ADR 0075 §2.1: the table's host, visible to everyone. It is a
       seat ROLE — who may change the house rules and spawn — and is
       unrelated to the monarch, whose crown sits in the marker column
       on the avatar row. Hence a chip next to the name, where the BOT
       chip lives, rather than a second crown: two crowns on one seat
       meaning two different things is exactly the confusion to avoid,
       and the glyph is kept small and paired with the word. -->
  {#if seat.is_host}
    <span
      class="tag host"
      title="Table host — sets this table's house rules, alongside the server admin"
    >
      <Icon name="crown" size={9} /> host
    </span>
  {/if}
  {#if isBot}
    <span class="tag bot" title={botTitle}>{botThinking ? "thinking…" : botLabel}</span>
  {/if}

  <div class="core-row">
    <!-- Left: mana pool floats alongside the avatar instead of stacking
         below it. Empty when no mana is pooled, so the column
         collapses and the avatar stays visually centred. -->
    <div class="side left" aria-hidden={!seat.mana_pool || seat.mana_pool.length === 0}>
      <ManaPoolPips pool={seat.mana_pool} />
    </div>

    <!-- svelte-ignore a11y_no_noninteractive_tabindex -->
    <div
      class="avatar-wrap"
      data-seat-id={seat.id}
      role={interactive ? "button" : "group"}
      tabindex={interactive ? 0 : undefined}
      onclick={handleAvatarClick}
      onkeydown={(e) => {
        if (interactive && (e.key === "Enter" || e.key === " ")) {
          e.preventDefault();
          handleAvatarClick();
        }
      }}
      aria-label={attackTargetable
        ? `attack ${displayLabel}`
        : `${displayLabel}, ${seat.life} life`}
    >
      {#if isBot}
        <span class="avatar bot-mark" aria-hidden="true">
          <Icon name="robot" size={30} />
        </span>
      {:else if avatar}
        {#key avatar}
          <img
            class="avatar"
            src={avatar}
            alt=""
            aria-hidden="true"
            onerror={() => (failedAvatarURL = avatar)}
          />
        {/key}
      {:else}
        <span class="avatar seat-dot-fallback" aria-hidden="true"></span>
      {/if}

      <!-- Life overlay sits on the bottom arc of the avatar circle. On
           self we expose ± chips flanking the number; on opponents it
           reads as a static chip. The keyword badges (#1201) sit
           beside it in the same row, rather than overlaying the
           avatar a second time. -->
      <div class="life-row">
        {#if isSelf}
          <span class="life-chip">
            <button
              type="button"
              class="life-btn dec"
              title="-1 life"
              aria-label="lose 1 life"
              onclick={(e) => {
                e.stopPropagation();
                changeLife(-1);
              }}>−</button
            >
            <span class="life">{seat.life}</span>
            <button
              type="button"
              class="life-btn inc"
              title="+1 life"
              aria-label="gain 1 life"
              onclick={(e) => {
                e.stopPropagation();
                changeLife(1);
              }}>+</button
            >
          </span>
        {:else}
          <span class="life-chip readonly">
            <span class="life">{seat.life}</span>
          </span>
        {/if}
        {#if keywordBadges.length > 0}
          <div class="seat-keywords" aria-label="keywords">
            {#each keywordBadges as badge (badge.key)}
              <span
                class="kw-badge"
                class:kw-icon={!!badge.icon}
                class:kw-text={!badge.icon}
                class:kw-protection={badge.kind === "protection"}
                title={badge.title}
                aria-label={badge.title}
              >
                {#if badge.icon}
                  <!-- eslint-disable-next-line svelte/no-at-html-tags -->
                  {@html badge.icon}
                {:else}
                  {badge.short}
                {/if}
              </span>
            {/each}
          </div>
        {/if}
      </div>

      <!-- Every life change the frame brought, oldest at the top, so
           two combat damage steps read as two numbers. Collapses to
           the net total past LIFE_POPUP_MAX_LINES; the life chip below
           always carries the resulting total either way. Visually
           hidden from the a11y tree — the live region below announces
           the whole group as one sentence, and it is always mounted so
           the announcement is reliable (the CombatArrows pattern). -->
      {#if popup}
        {#key popup.key}
          <span
            class="dmg-popup"
            class:stacked={popup.view.lines.length > 1}
            in:floatUp
            out:fadeOut
            aria-hidden="true"
          >
            {#each popup.view.lines as line, i (i)}
              <span class="dmg-line {line.tone}">{line.text}</span>
            {/each}
            {#if popup.view.collapsed}
              <span class="dmg-note">×{popup.view.count}</span>
            {/if}
            {#if popup.view.gap}
              <span class="dmg-note" title="older changes are no longer in the log">…</span>
            {/if}
          </span>
        {/key}
      {/if}
    </div>

    <!-- Right: monarch/initiative toggles + poison/energy steppers +
         any ad-hoc player counters. Stacked vertically so they don't
         push the panel taller; each row is the same height as a
         single chip. -->
    <div class="side right">
      {#if isSelf}
        <div class="crown-row">
          <button
            type="button"
            class="marker monarch"
            class:active={isMonarch}
            title={isMonarch ? "release the monarch" : "claim the monarch"}
            aria-label="toggle monarch"
            aria-pressed={isMonarch}
            onclick={(e) => {
              e.stopPropagation();
              toggleMonarch();
            }}><Icon name="crown" size={13} /></button
          >
          <button
            type="button"
            class="marker initiative"
            class:active={isInitiative}
            title={isInitiative ? "release the initiative" : "claim the initiative"}
            aria-label="toggle initiative"
            aria-pressed={isInitiative}
            onclick={(e) => {
              e.stopPropagation();
              toggleInitiative();
            }}><Icon name="sword" size={13} /></button
          >
        </div>
        <span class="counter-inline poison" title="poison counters">
          <span class="counter-icon" aria-hidden="true"><Icon name="drop" size={11} /></span>
          <button
            type="button"
            class="counter-btn"
            aria-label="lose 1 poison"
            onclick={(e) => {
              e.stopPropagation();
              changePoison(-1);
            }}>−</button
          >
          <span class="counter-val">{seat.poison ?? 0}</span>
          <button
            type="button"
            class="counter-btn"
            aria-label="gain 1 poison"
            onclick={(e) => {
              e.stopPropagation();
              changePoison(1);
            }}>+</button
          >
        </span>
        <span class="counter-inline energy" title="energy counters">
          <span class="counter-icon" aria-hidden="true"><Icon name="bolt" size={11} /></span>
          <button
            type="button"
            class="counter-btn"
            aria-label="lose 1 energy"
            onclick={(e) => {
              e.stopPropagation();
              changeEnergy(-1);
            }}>−</button
          >
          <span class="counter-val">{seat.energy ?? 0}</span>
          <button
            type="button"
            class="counter-btn"
            aria-label="gain 1 energy"
            onclick={(e) => {
              e.stopPropagation();
              changeEnergy(1);
            }}>+</button
          >
        </span>
      {:else}
        {#if isMonarch}
          <span class="marker monarch active" title="monarch" aria-label="monarch"
            ><Icon name="crown" size={12} /></span
          >
        {/if}
        {#if isInitiative}
          <span class="marker initiative active" title="initiative" aria-label="initiative"
            ><Icon name="sword" size={12} /></span
          >
        {/if}
        {#if (seat.poison ?? 0) > 0}
          <span class="marker poison" title={`${seat.poison} poison`} aria-label="poison">
            <Icon name="drop" size={11} />{seat.poison}
          </span>
        {/if}
        {#if (seat.energy ?? 0) > 0}
          <span class="marker energy" title={`${seat.energy} energy`} aria-label="energy">
            <Icon name="bolt" size={11} />{seat.energy}
          </span>
        {/if}
        {#if seat.counters}
          {#each Object.entries(seat.counters) as [name, count] (name)}
            {#if count > 0 && name !== "poison" && name !== "energy"}
              <span class="marker counter" title={`${count} ${name}`} aria-label={name}>
                {#if name === "experience"}<Icon
                    name="star"
                    size={11}
                  />{:else if name === "rad"}<Icon name="rad" size={11} />{:else}<Icon
                    name="dot"
                    size={11}
                  />{/if}{count}
              </span>
            {/if}
          {/each}
        {/if}
      {/if}
    </div>
  </div>

  <!-- Emblems (CR 114, #623). They sit under the core row rather
       than in the marker column because an emblem is a sentence, not
       a count: the label is what identifies it and the printed
       ability is the hover. Shown on every seat, self and opponent
       alike — emblems are public. -->
  {#if seat.emblems && seat.emblems.length > 0}
    <div class="emblems" aria-label="emblems">
      {#each seat.emblems as emblem (emblem.instance_id)}
        <span class="emblem" title={`${emblem.label} — ${emblem.text}`}>
          <Icon name="star" size={11} />{emblem.label}
        </span>
      {/each}
    </div>
  {/if}

  {#if seat.eliminated}
    <span class="tag elim">eliminated</span>
  {/if}

  <!-- The life change as words. Always mounted and empty between
       changes, so a screen reader announces each group as it lands
       rather than missing a region that appeared with its text
       already in it. S11.5: this is text, not motion — it says the
       same thing with animations off and reduce-motion on. -->
  <span class="sr-only" role="status" aria-live="polite" aria-atomic="true"
    >{popup ? popup.view.label : ""}</span
  >
</div>

<style>
  .identity {
    --avatar-size: calc(var(--avatar-size-base, 88px) * var(--card-scale, 1));
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
    color: var(--fg);
    font-size: 12px;
    position: relative;
    min-width: 0;
    max-width: 100%;
  }
  .name {
    font-weight: 600;
    letter-spacing: 0.01em;
    max-width: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    color: var(--fg);
    text-shadow: 0 1px 0 rgba(0, 0, 0, 0.4);
  }
  .avatar-wrap {
    position: relative;
    width: var(--avatar-size);
    height: var(--avatar-size);
    border-radius: 50%;
    display: grid;
    place-items: center;
    padding: 3px;
    background: var(--surface-raised);
    border: 2.5px solid color-mix(in srgb, var(--seat-color, #888) 55%, var(--surface));
    box-shadow: 0 8px 20px rgba(0, 0, 0, 0.5);
    transition:
      box-shadow 160ms var(--ease),
      border-color 160ms var(--ease);
    box-sizing: border-box;
  }
  .avatar {
    width: 100%;
    height: 100%;
    border-radius: 50%;
    object-fit: cover;
    display: block;
  }
  .seat-dot-fallback {
    background: var(--seat-color, #888);
  }
  /* Bot seats get a glyph rather than a portrait: a bot has no
     Discord avatar, and falling through to its commander's art crop
     would read as a human who happens to play that commander. */
  .bot-mark {
    display: grid;
    place-items: center;
    background: color-mix(in srgb, var(--seat-color, #888) 22%, var(--surface-raised));
    color: color-mix(in srgb, var(--seat-color, #888) 70%, var(--fg));
  }
  .identity.bot .avatar-wrap {
    border-style: dashed;
  }
  /* The thinking pulse. Only ever applied when the animations setting
     and the reduce-motion preference both allow it (see
     animateThinking) — the chip's "thinking…" text carries the
     information on its own. The @media guard is the belt to that's
     braces for a browser-level preference set after load. */
  .identity.thinking .avatar-wrap {
    animation: bot-think 1.4s ease-in-out infinite;
  }
  @keyframes bot-think {
    0%,
    100% {
      box-shadow:
        0 0 0 2px var(--gold),
        0 0 12px rgba(255, 208, 122, 0.35);
    }
    50% {
      box-shadow:
        0 0 0 3px var(--gold),
        0 0 30px rgba(255, 208, 122, 0.8);
    }
  }
  @media (prefers-reduced-motion: reduce) {
    .identity.thinking .avatar-wrap {
      animation: none;
    }
  }

  /* State rings translate the old pill shadows into circle-shaped ones.
     Active = whose turn it is → seat-coloured ring.
     Priority = holds priority right now → gold ring (overrides active). */
  .identity.active .avatar-wrap {
    border-color: var(--seat-color, #5fb0ff);
    box-shadow:
      0 0 0 1px var(--seat-color, #5fb0ff),
      0 0 22px color-mix(in srgb, var(--seat-color, #5fb0ff) 45%, transparent),
      inset 0 1px 0 rgba(255, 255, 255, 0.08);
  }
  .identity.priority .avatar-wrap {
    border-color: var(--gold);
    box-shadow:
      0 0 0 2px var(--gold),
      0 0 26px rgba(255, 208, 122, 0.65),
      inset 0 1px 0 rgba(255, 255, 255, 0.08);
  }
  .identity.targetable .avatar-wrap {
    cursor: pointer;
    border-color: var(--danger);
    box-shadow:
      0 0 0 2px var(--danger),
      0 0 22px rgba(255, 122, 122, 0.55);
  }
  .identity.targetable .avatar-wrap:hover {
    box-shadow:
      0 0 0 3px var(--danger),
      0 0 28px rgba(255, 122, 122, 0.75);
  }
  .identity.cast-targetable .avatar-wrap {
    cursor: pointer;
    border-color: var(--gold);
    box-shadow:
      0 0 0 2px var(--gold),
      0 0 22px rgba(255, 208, 122, 0.55);
  }
  .identity.cast-targetable .avatar-wrap:hover {
    box-shadow:
      0 0 0 3px var(--gold),
      0 0 28px rgba(255, 208, 122, 0.75);
  }
  .identity.cast-picked .avatar-wrap {
    border-color: #ffe69a;
    box-shadow:
      0 0 0 3px #ffe69a,
      0 0 28px rgba(255, 230, 154, 0.8);
  }
  .identity.eliminated {
    opacity: 0.5;
    filter: grayscale(0.6);
  }

  /* Life row overlays the bottom third of the avatar circle: the life
     chip itself, plus (#1201) any keyword badges beside it. Anchored
     as a row rather than each child positioning itself, so the
     badges sit flush against the chip instead of needing their own
     copy of this math. */
  .life-row {
    position: absolute;
    left: 50%;
    bottom: -6px;
    transform: translateX(-50%);
    display: flex;
    align-items: center;
    gap: 4px;
  }
  /* Ellipse backdrop sits flush with the circle's lower arc so the
     number reads as part of the identity disc rather than a floating
     badge. */
  .life-chip {
    display: inline-flex;
    align-items: center;
    gap: 2px;
    padding: 2px 6px;
    border-radius: 999px;
    background:
      linear-gradient(180deg, rgba(255, 255, 255, 0.08) 0%, rgba(255, 255, 255, 0) 60%),
      rgba(6, 10, 22, 0.92);
    border: 1px solid rgba(255, 255, 255, 0.1);
    box-shadow: 0 3px 10px rgba(0, 0, 0, 0.5);
  }
  .life-chip.readonly {
    padding: 2px 10px;
  }
  /* #1201: a seat's hexproof / protection badges, in the same visual
     language as the card protection badge (#979,
     KeywordBadgeRow.svelte) — same shell, same icon for hexproof,
     same shield tint for protection. Not a shared component: the
     card row is absolutely positioned across a card's bottom edge,
     which has nothing to do with sitting beside a life chip, so only
     the class names and their declarations are reused here. */
  .seat-keywords {
    display: flex;
    align-items: center;
    gap: 2px;
  }
  .kw-badge {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    color: #fff;
    background: rgba(0, 0, 0, 0.72);
    border: 1px solid rgba(255, 255, 255, 0.25);
    border-radius: 3px;
    backdrop-filter: blur(6px);
    -webkit-backdrop-filter: blur(6px);
    line-height: 1;
  }
  .kw-badge.kw-icon {
    width: 14px;
    height: 14px;
    padding: 1px;
  }
  .kw-badge.kw-icon :global(svg) {
    width: 100%;
    height: 100%;
    display: block;
  }
  .kw-badge.kw-text {
    padding: 1px 3px;
    font-size: 9px;
    font-weight: 700;
    letter-spacing: 0.04em;
    text-shadow: 0 1px 0 rgba(0, 0, 0, 0.6);
    min-width: 14px;
    text-align: center;
  }
  /* Protection is a shield rather than an ability the seat uses, so
     it reads as a different thing — same tint KeywordBadgeRow gives
     a card's protection badge. */
  .kw-badge.kw-protection {
    background: rgba(24, 48, 92, 0.85);
    border-color: rgba(160, 200, 255, 0.45);
  }
  .life {
    font-weight: 800;
    font-size: 15px;
    min-width: 22px;
    text-align: center;
    font-variant-numeric: tabular-nums;
    color: var(--fg);
    letter-spacing: -0.02em;
    text-shadow: 0 1px 0 rgba(0, 0, 0, 0.5);
  }
  .life-btn {
    width: 18px;
    height: 18px;
    padding: 0;
    border-radius: 50%;
    border: 1px solid rgba(255, 255, 255, 0.1);
    background: rgba(0, 0, 0, 0.3);
    color: var(--fg);
    font-size: 13px;
    line-height: 1;
    cursor: pointer;
    font-family: inherit;
    box-shadow: none;
    transition:
      border-color 120ms var(--ease),
      background 120ms var(--ease),
      color 120ms var(--ease);
  }
  .life-btn.dec:hover {
    background: rgba(255, 122, 122, 0.18);
    border-color: rgba(255, 122, 122, 0.5);
    color: var(--danger);
  }
  .life-btn.inc:hover {
    background: rgba(122, 255, 154, 0.18);
    border-color: rgba(122, 255, 154, 0.5);
    color: var(--mint);
  }

  /* Damage popup floats up over the circle's top arc; heal/gain
     intentionally also floats up (same pattern as the old header).
     A column, because one frame can bring several changes (#703):
     oldest at the top, newest nearest the life chip it just moved.
     The stack grows upward from a fixed bottom edge so the newest
     number never jumps. */
  .dmg-popup {
    position: absolute;
    left: 50%;
    /* Anchored where the old single-number popup sat (top: -14px on a
       ~29px line), so one change looks exactly as it did. */
    bottom: calc(100% - 15px);
    transform: translateX(-50%);
    display: flex;
    flex-direction: column;
    align-items: center;
    line-height: 1.05;
    font-size: 24px;
    font-weight: 800;
    pointer-events: none;
    z-index: 20;
    text-shadow:
      0 1px 0 rgba(0, 0, 0, 0.85),
      0 0 8px rgba(0, 0, 0, 0.6);
  }
  /* A stack steps down one size — four 24px numbers are taller than
     the avatar they sit over — and stays one size within itself: no
     delta in the group is more important than the others. */
  .dmg-popup.stacked .dmg-line {
    font-size: 19px;
  }
  .dmg-line.loss {
    color: #ff7a7a;
  }
  .dmg-line.gain {
    color: #7aff9a;
  }
  /* The "×7" on a collapsed run, and the "…" that says older changes
     have rolled off the server's log. */
  .dmg-note {
    font-size: 11px;
    font-weight: 700;
    letter-spacing: 0.04em;
    color: var(--fg-dim);
  }
  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
    border: 0;
  }

  /* Horizontal arrangement: mana (left column) — avatar (center) —
     counters (right column). The side columns stack vertically so a
     full marker set (crown + sword + poison + energy) is roughly the
     same height as the avatar, keeping the panel compact. */
  .core-row {
    /* In the rail everything stacks: avatar, then the floating mana
       pool, then the marker chips as a wrapping row. */
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 6px;
    max-width: 100%;
  }
  .side {
    display: flex;
    flex-direction: row;
    flex-wrap: wrap;
    justify-content: center;
    gap: 4px;
    flex: 0 0 auto;
    min-width: 0;
    max-width: 100%;
  }
  .side.left {
    order: 1;
  }
  .side.left[aria-hidden="true"] {
    display: none;
  }
  .side.right {
    order: 2;
    margin-top: 6px;
  }
  /* Crown + sword sit on a single row so monarch/initiative read as
     peer toggles rather than a tall stack. */
  .crown-row {
    display: flex;
    gap: 4px;
  }
  .marker {
    font-size: 12px;
    line-height: 1;
    display: inline-flex;
    align-items: center;
    gap: 2px;
    padding: 2px 6px;
    border-radius: 999px;
    background: rgba(0, 0, 0, 0.3);
    color: #c8c8c8;
    border: 1px solid rgba(255, 255, 255, 0.05);
    cursor: default;
    backdrop-filter: blur(4px);
  }
  button.marker {
    font-family: inherit;
    cursor: pointer;
    box-shadow: none;
    transition:
      transform 120ms var(--ease),
      border-color 120ms var(--ease),
      background 120ms var(--ease);
  }
  button.marker:hover {
    border-color: rgba(255, 255, 255, 0.2);
    background: rgba(0, 0, 0, 0.55);
  }
  button.marker:active {
    transform: scale(0.95);
  }
  .marker.active.monarch {
    color: var(--gold);
    border-color: rgba(255, 208, 122, 0.7);
    background: var(--gold-soft);
    box-shadow: 0 0 10px rgba(255, 208, 122, 0.35);
  }
  .marker.active.initiative {
    color: #b08aff;
    border-color: rgba(176, 138, 255, 0.7);
    background: rgba(176, 138, 255, 0.18);
    box-shadow: 0 0 10px rgba(176, 138, 255, 0.35);
  }
  .marker.poison {
    color: var(--mint);
  }
  .marker.energy {
    color: var(--gold);
  }
  .marker.counter {
    color: #b8c8e8;
  }
  .counter-inline {
    display: inline-flex;
    align-items: center;
    gap: 2px;
    padding: 2px 6px;
    border-radius: 999px;
    background: rgba(0, 0, 0, 0.3);
    border: 1px solid rgba(255, 255, 255, 0.06);
  }
  .counter-icon {
    font-size: 11px;
    line-height: 1;
    margin-right: 2px;
  }
  .counter-val {
    min-width: 14px;
    text-align: center;
    font-weight: 700;
    font-variant-numeric: tabular-nums;
    color: var(--fg);
  }
  .counter-inline.poison .counter-val {
    color: var(--mint);
  }
  .counter-inline.energy .counter-val {
    color: var(--gold);
  }
  .counter-btn {
    width: 14px;
    height: 14px;
    padding: 0;
    border: 0;
    background: transparent;
    color: var(--fg-dim);
    font-size: 11px;
    line-height: 1;
    cursor: pointer;
    font-family: inherit;
    border-radius: 50%;
    box-shadow: none;
    transition:
      background 100ms var(--ease),
      color 100ms var(--ease);
  }
  .counter-btn:hover {
    background: rgba(255, 255, 255, 0.08);
    color: var(--fg);
  }

  /* Emblem chips (#623). A wrapping row so several fit at any panel
     width, each truncated to the panel rather than widening it — the
     full text is the title attribute. */
  .emblems {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: 4px;
    margin-top: 6px;
    max-width: 100%;
  }
  .emblem {
    font-size: 11px;
    line-height: 1;
    display: inline-flex;
    align-items: center;
    gap: 3px;
    padding: 3px 7px;
    border-radius: 999px;
    background: var(--gold-soft, rgba(255, 208, 122, 0.14));
    border: 1px solid rgba(255, 208, 122, 0.45);
    color: var(--gold);
    max-width: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    cursor: default;
  }

  .tag {
    font-size: 9px;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    padding: 2px 7px;
    border-radius: 999px;
    background: rgba(0, 0, 0, 0.35);
    border: 1px solid rgba(255, 255, 255, 0.06);
    font-weight: 700;
    margin-top: 6px;
  }
  .tag.elim {
    color: var(--danger);
    border-color: rgba(255, 122, 122, 0.4);
  }
  .tag.host {
    display: inline-flex;
    align-items: center;
    gap: 3px;
    margin-top: 0;
    margin-bottom: 2px;
    color: var(--gold, #ffd07a);
    border-color: rgba(255, 208, 122, 0.4);
    background: rgba(255, 208, 122, 0.12);
  }
  .tag.bot {
    margin-top: 0;
    margin-bottom: 2px;
    color: color-mix(in srgb, var(--seat-color, #888) 60%, var(--fg));
    border-color: color-mix(in srgb, var(--seat-color, #888) 45%, transparent);
    background: rgba(0, 0, 0, 0.28);
  }
</style>
