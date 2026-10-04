# ADR 0118 — Strict payment by default, Cast anyway, and alternative costs for every spell

**Status:** Accepted (owner answers 2026-10-04) · 2026-10-04 · S59 — Automated table: clicks that act, payment that counts (tracker [#2189](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2189))
**Issues:** [#2188](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2188) (strict payment with auto-tap by default, and the Cast anyway row) and [#2163](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2163) (alternative costs for every spell you cast: Fist of Suns, Jodah, Archmage Eternal, Leyline of Mutation, Omniscience).
**Owner decisions:** the three answers of 2026-10-04 recorded on the two issues, and four more answers given in review of this ADR the same day, all quoted under [Owner decisions](#owner-decisions-2026-10-04). They are binding. The review answers settled four of the calls the first draft made: the log wording, the settings migration, a confirmation before an unpaid cast, and the practice table. The six calls left are listed under [Calls made here](#calls-made-here), and the owner can still overturn any of them.
**Numbering:** checked with the AGENTS.md §4 sweep on 2026-10-04. I ran `git fetch --prune` and listed `docs/decisions/` on all 39 remote heads. 0117 and 0120 are on `origin/develop`, 0119 is on `origin/docs/s60-adr-0119-stack-you-can-follow` (in review), and no branch has 0118. 0118 has been reserved for #2188 since 2026-10-04 (ADR 0117's header), so this ADR takes **0118**.
**Amends:** [ADR 0011](0011-mana-pool-and-auto-tapper.md) §1 (strict is the default; permissive stays an opt-out per client) and §7 (the planner tops up a partly funded pool). [ADR 0076](0076-tutorial.md) §2.1 step 1's copy and §2.2's forced `strictMana` value. [ADR 0111](0111-action-dock.md) Delivery PR 4's insufficient-mana request (its buttons when the refused cast was already auto-tapped).
**Builds on:** [ADR 0040](0040-mana-pipeline.md) (what may pay), [ADR 0066](0066-granted-cast-and-play-permissions.md) (who may cast from where, and at what price), [ADR 0068](0068-the-mana-spent-on-a-spell.md) (`OnPaper`), [ADR 0074](0074-triggered-mana-abilities.md) (triggered mana, unchanged), [ADR 0105](0105-legal-action-highlights.md) (the ready ring reads the server's move list), [ADR 0109](0109-rule-gates-land-types-mana-and-cost-components.md) (cost components), [ADR 0117](0117-click-to-act-and-a-per-colour-mana-stepper.md) §3 (the popover's Sandbox section).

This ADR was written plan-first. No code changed with it. The changes land in the PRs listed under [Delivery](#delivery).

---

## Context

ADR 0011 §1 made the S15 mana gate opt-in: `gameplay.strictMana` defaults to off, and a cast with neither `strict` nor `auto_tap` takes the sandbox posture, where the engine waives the mana and marks the payment `OnPaper`. The table is automated now. Bots, drag casts, special actions and attacks already pay through the auto-tapper, and ADR 0117 made a left-click do what the card does. A spell's cost should be paid too.

Two things stand in the way. A strict table has no answer for a cost the engine cannot offer yet: Jodah, Archmage Eternal's {W}{U}{B}{R}{G} for every spell is the owner's example. And the escape that exists, the dock's "Cast anyway", only appears after a refused cast. The owner's answer is a standing escape on every castable card, and the real fix for Jodah in the same sprint.

Every claim below was checked on `origin/develop` at `8a838fe1`. Every rule was checked against the pinned Comprehensive Rules (`MagicCompRules 20260925.txt`).

### Owner decisions (2026-10-04)

1. **Strict payment with auto-tap becomes the default** (#2188: "Yes, in a separate ADR").
2. **Cast anyway from the right-click menu** (#2188, owner additions): "Every card the viewer could cast (hand, exile strip, commander) gets a 'Cast anyway (don't pay)' row whenever strict payment is on. It casts with `force_cast`, and the game log says '<player> cast <card> without paying its cost' so the table sees it." The row is always offered, not only when the viewer is short.
3. **#2163 is built now, in S59** (#2163, claimed 2026-10-04): "built alongside ADR 0118 … so Jodah, Archmage Eternal's {W}{U}{B}{R}{G} is a real offer the auto-tapper prices." The seam is a standing static that adds an offer, priced and timed like a printed one, to each spell its controller casts from the named zone, read by the cast path, the enumerator, the auto-tapper and the client. Omniscience's {0} keeps each spell's normal timing. Leyline of Mutation's opening-hand clause ships as a caveat, as the other Leylines' does.

The owner answered the first draft's calls in review on 2026-10-04 (PR #2221):

4. **The log wording** is "<player> cast <card> without paying its mana cost", not "… its cost". This replaces the wording quoted in answer 2.
5. **The migration** is confirmed: settings v16 moves everyone to strict once, synced copies included, and a later opt-out sticks.
6. **Cast anyway asks first.** The row opens a small confirmation dialog, "Cast <card> without paying its mana cost?", with Cast and Cancel. The row stays in the Sandbox section, and the server still judges timing.
7. **The practice table** is confirmed: it forces strict on.

The owner has `strictMana` on in their own synced settings today; on `develop` it still defaults to off.

### The rules

- **CR 118.9:** "An alternative cost is a cost listed in a spell's text, or applied to it from another effect, that its controller may pay rather than paying the spell's mana cost." Fist of Suns, Jodah and Leyline of Mutation apply one ("You may pay {W}{U}{B}{R}{G} rather than pay the mana cost for spells you cast"). So does Omniscience: the rule's second phrasing is "You may cast [this object] without paying its mana cost."
- **CR 118.9a:** "Only one alternative cost can be applied to any one spell as it's being cast." **CR 601.2b** says the same of methods: "A player can't apply two alternative methods of casting or two alternative costs to a single spell." Flashback (CR 702.34a) and escape (CR 702.138a) are alternative costs, so a card cast by one of them cannot also take Jodah's.
- **CR 118.9c:** an alternative cost "doesn't change a spell's mana cost, only what its controller has to pay to cast it."
- **CR 118.9d:** "any additional costs, cost increases, and cost reductions that affect that spell are applied to that alternative cost." Commander tax is an additional cost (CR 903.8), so a commander cast for {W}{U}{B}{R}{G} from the command zone still pays the tax.
- **CR 118.6 and 118.6a:** a spell with no mana cost has an unpayable cost, but "If an alternative cost is applied to an unpayable cost, including an effect that allows a player to cast a spell without paying its mana cost, the alternative cost may be paid." Omniscience and Jodah can cast Ancestral Vision from hand.
- **CR 107.3b:** casting a spell "while paying neither its mana cost nor an alternative cost that includes X" fixes X at 0. That covers both {W}{U}{B}{R}{G} and Omniscience's free cast.
- **CR 601.2f:** the total cost "is the mana cost or alternative cost (as determined in rule 601.2b), plus all additional costs and cost increases, and minus all cost reductions."
- **CR 601.2g and 601.2h:** mana abilities are activated before costs are paid, and then "The player pays the total cost … Partial payments are not allowed." The auto-tapper is the engine doing 601.2g for the player.
- **CR 117.1a and 307.1:** the timing of a cast is the card's own. Nothing in Omniscience's text changes it.
- **CR 103.6a:** "If a card allows a player to begin the game with that card on the battlefield, the player taking this action puts that card onto the battlefield." The engine has no such step (Leyline of Anticipation's caveat).
- **CR 106.1b:** the six types of mana. The auto-tapper colours {W}{U}{B}{R}{G} from them.

### What exists

**The payment postures.** One setting and a handful of hard-wired stamps decide how a payment is made:

| Path | `strictMana` off | `strictMana` on |
|---|---|---|
| Click cast (hand, exile strip, commander) | `strict: false`: the mana is waived, `OnPaper` | `strict: true`: the pool alone must cover it, or the cast is refused with `insufficient_mana` |
| Drag cast (#1508, #1622, #2207) | `strict: true, auto_tap: true` | the same |
| Catalog activation (#1296) | nothing stamped: waived | `strict: true, auto_tap: true` |
| Special actions, Room doors, attack tax, library top | `strict: true, auto_tap: true` always (`contextMenu.logic.ts:1690`, `roomDoors.ts:70`, `contextMenu.logic.ts:1482`, `libraryTop.ts:128`) | the same |
| Bots | every move is `strict` + `auto_tap` (`legal/cast.go:14-16`) | the same |

The client stamps casts and activations in one place: `sendAction` in `routes/Game.svelte:522-543` calls `stampManaEnforcement` (`lib/manaEnforcement.ts:23-36`), which leaves a payload alone if it already says `strict` (`:28`) and otherwise writes `strict: strictMana` on a cast. A dragged cast gets its flags from `applyCastChoices` (`lib/targeting.ts:220-223`). The server's default for a payload with neither flag is permissive (`game/mutations.go:495-503`, `CastSpellParams.Strict`).

**The setting.** `gameplay.strictMana` is declared at `lib/settings.ts:162`, defaults to `false` at `:382` and is `"synced"` at `:508`. `SETTINGS_VERSION` is 15 (`:282`). The v3→v4 note (`:669-673`) says the shallow merge fills the field from defaults. The v14→v15 block (`:776-786`) is the precedent for forcing a value on everyone once: ADR 0105 owner decision 4 turned `highlightLegalActions` on for existing players, and from v15 on the stored choice is honoured. An account copy goes through the same `migrate` with the version that wrote it (`applySyncedCopy`, `:589-607`), so a migration reaches synced settings as well as the browser's own. The Settings row is labelled "Strict mana enforcement", and its help says "Default is off" (`components/Settings.svelte:918-931`).

**What strict does on the server.** `CastSpell` runs `applyAutoTapLocked` when `AutoTap` is set (`mutations.go:1400-1404`), then `applyCastCostLocked` (`:1413`). That has three outcomes (`:1738-1830`): strict and payable spends the pool; strict, short and not forced returns `*InsufficientManaError`; permissive or `ForceCast` leaves the pool untouched, emits `EventCostWarning` and marks the payment `OnPaper` (`:1800-1820`). The life half of a Phyrexian cost is still paid in every mode. `ForceCast` "overrides the Strict gate for this one cast" (`:505-511`).

**A gap in the auto-tapper.** `applyAutoTapLocked` skips planning when the pool already covers the cost (`mutations.go:1932`). When the pool covers only part of it, the planner is asked for the whole cost (`:1950` passes `cost`, not what the pool is missing), and `gatherTapSources` reads only permanents and hand sources, never the pool. The enumerator's affordability check has the same shape: the pool covers it, or the sources cover all of it (`legal/cast.go:1556-1575`). So a player with {G} floating and one untapped Forest is told a {1}{G} creature is not castable, and if they cast it anyway with `auto_tap` the cast is refused. Under the sandbox default this rarely mattered. Under strict by default it bites everyone who taps a land before casting, which the tutorial teaches in step 5.

**The insufficient-mana request.** When a strict cast is refused, `Game.svelte:554-565` holds the card and its missing symbols, and the dock shows `insufficientManaRequest` (`lib/targetingDock.ts:197-238`): "Auto-tap & cast" as the primary (it opens `AutoTapPreviewModal`, which confirms before tapping), "Cancel" and "Cast anyway" as secondaries. Cast anyway has no key, on purpose (`targetingDock.test.ts:167`). Both retries replay the stashed payload (`lastCastByCardID`, `Game.svelte:521`), so targets, modes and X survive: `castAnyway` adds `strict: true, force_cast: true` (`:581-591`), and `confirmAutoTap` adds `auto_tap` and `locked_sources` (`:616-628`). Cast anyway's title says "cast it without paying the missing mana", but `ForceCast` pays none of it: the pool is left alone.

**What a hand card's right-click opens.** `Card.svelte` `handleContextMenu` (`:554-567`) opens the override menu with admin overrides on, and otherwise the light popover (`ManaAbilityMenu`) when `hasMenu` (`:343-349`) is true. For a hand card that is true only when it has `zone_abilities`, `zone_mana_abilities` or special actions, which `Hand.svelte:683-706` wires for the viewer's own cards. A plain spell in hand has no menu: the right-click falls through to the browser's own menu. The exile strip's cards (`ExileStrip.svelte:467-474`) and the commanders in it (#2207) wire only `onClick`. The command-zone panel's card (`CommandZone.svelte:202-211`) wires `onActivateAbility` for its `zone_abilities`. The popover already has a Sandbox section at the bottom, `role="group"` named `sandbox`, holding Tap or Untap on a permanent (ADR 0117 §3, `ManaAbilityMenu.svelte:324-345`). The override menu's hand section lists special actions, then "move to" (`contextMenu.logic.ts:1768-1785`).

**What the hand lets you click.** `canCastFromHand` (`lib/timing.ts:237`) is LEGAL as soon as the server's move list has a cast for the card (`:243-244`). The move list is strict and auto-tapped (ADR 0105, `legal/cast.go`), so whenever the seat owes a decision, a card the board cannot pay for is dimmed and its click is not wired (`Hand.svelte:705`), whatever `strictMana` says. Its later branches only find the sentence for a denial: `cant_cast`, targets, modes, discard, sacrifice, an either/or branch, and finally "Only at sorcery speed", which the comment calls a hint that can be wrong when the real reason is mana (`:403-409`).

**The log.** `EventCast` (`mutations.go:1704-1711`) carries the actor, the card and the zone it left. `projectEvent` turns it into `LogCast` (`protocol/log.go:916-925`), and `renderLogText` writes "P cast C" or "P cast C from Z" (`:1814-1818`) on the server. `EventCostWarning` is classed `silentServerDiagnostic` (`protocol/log_event_kind_gate_test.go:106`) and is emitted for a permissive cast and a forced cast alike, so nothing on the wire tells the two apart. `Event` is part of the snapshot (`game/snapshot.go:382`, recorded in `testdata/snapshot_shape/v7.txt:388-436`; `SnapshotSchemaVersion` is 7, `snapshot.go:208`).

**The tutorial.** The practice table forces four settings, `strictMana` off among them (`lib/practiceTable.ts:8`, `FORCED_SETTINGS` at `:66`), and settingsSync never uploads a forced value. Step 1's copy says "You can cast whatever your lands could pay for without tapping them first" (`lib/tutorialSteps.ts:164`), and the file header's last note says strictMana off still dims what the move list leaves out (`:35-38`). Step 5 asks the player to tap a Forest, and step 6 to cast a creature from the lit cards. No step relies on an unpaid cast: step 6's `cannot` already gives up when the move list offers no creature (`:259-272`). ADR 0076's Context and §2.2 describe the forced value (`0076-tutorial.md:55-57`, `:127-130`).

**The e2e suite.** No spec casts a spell through the client. The boards are staged through the admin socket (`move_card`, `draw_card`), and the one spec that reads cast legality, `legal-highlights-1789.spec.ts`, reads the server's digest, which is strict already. The S19 helper seeds `{ gameplay: { alwaysStopOpponentStack: true } }` with no `__version` (`s19-helpers.ts:634-643`), which the migration below will turn strict; it casts nothing. No spec selects "Cast anyway" or "Auto-tap & cast".

**Alternative costs today.** `CastOffersForLocked` (`game/cast_zones.go:264-296`) lists what a cast out of a zone may claim: `nil` for the printed cost when `validateCastPathLocked` allows a claim of nothing, then the card's own offers (`AlternativeCostsOfferedFromZone`), then the one the cast's permission synthesises (`CastPermission.AlternativeCostFor`, `cast_permission.go:655-685`), then other stored permissions over the same card (#1729). Each is filtered by `validateCastPathLocked` (`cast_zones.go:350-416`) and `AlternativeCostPayableLocked`. Its readers are the view (`protocol/view.go:5022-5025`, then `viewOfAlternativeCosts` at `:5645`), the enumerator (`legal/cast.go:262`, `:278`) and the commander-return check (`game/commander_return.go:185`). A claimed key is resolved by `resolveAlternativeCostLocked` (`alternative_cost.go:465-486`): the card's own offer, a disturb front face, then the permission's. Its callers are `CastSpell` (`mutations.go:755`) and `effectiveCostLocked` (`:2777`), which is what the price, the auto-tap preview and the auto-tapper read. `CastCostFor` (`cast_cost.go:49-58`) lets a permission's flat `Cost` (cascade's {0}, airbend's {2}) win over any offer. CR 118.6a is already honoured: `ErrNoManaCost` is returned only when the cast pays the printed cost (`mutations.go:814-828`, `castPaysPrintedCost` in `no_mana_cost.go:59`).

No battlefield static adds an offer to every spell. Standing permissions are derived from the battlefield on every query (`standingCastPermissionsLocked`, `cast_permission.go:1270`), but a permission is a reason a cast is legal from a zone, not an extra price, and `storedHandPermissionLocked` (`:1069-1083`) skips every standing permission over a hand. The Bringers' {W}{U}{B}{R}{G} is the card's own offer (`effects/bringer_alternative_cost.go`, key `bringer-wubrg`). A claimed key lands on `StackItem.AltCost` (`stack.go:349-358`, persisted as `altCost`, `snapshot.go:1104`). The resolution readers look it up with `AlternativeCostByKey` (`alternative_cost.go:898`, `:943`, `:1020`, `modes.go:441`) and do nothing when the card does not print it. The registry row is `granted-alternative-costs` (`roadmap/registry.go:2548`), waiting on the four cards.

---

## Decision

### 1. Strict payment with auto-tap is the default

**The setting.** `gameplay.strictMana` defaults to `true`. The field, its name and its synced scope are unchanged. Off stays a supported choice: the sandbox posture for a table that tracks mana on paper (ADR 0011 §1).

**A clicked cast auto-taps.** With `strictMana` on, `stampManaEnforcement` writes `strict: true, auto_tap: true` on a cast, as it already does on an activation (#1296) and as a drag already does. A click on a card the board can pay for therefore taps the lands and casts, in one click. With it off, a cast is stamped `strict: false` exactly as today, byte for byte. A payload that already says `strict` is still left alone, so the dock's retries and the Cast anyway row keep their own flags.

**The pool tops up** (amends ADR 0011 §7). The auto-tapper plans only what the floating pool is missing, so mana already in the pool is spent first and the plan pays the rest. The enumerator's affordability check (`canPayExcluding`) asks the same question through the same function, so the ready ring, the click and the engine agree: {G} floating plus one untapped Forest casts a {1}{G} creature. Every auto-tap payer gets it: casts, activations, special actions and the attack tax. The planner remains read-only, and the plan still materialises atomically under the cast's write lock. This is a server change (PR 3) and goes in before the default flips.

**The insufficient-mana request** (amends ADR 0111 PR 4). A refused cast whose payload already had `auto_tap` drops the "Auto-tap & cast" button: the planner found no plan, and the preview would find none either. Cancel becomes the primary and Cast anyway stays a secondary with no key. A refused cast without `auto_tap` keeps today's three buttons. After this ADR no client surface sends that shape, but the server still accepts it, and the request reads the payload rather than the setting. Cast anyway's title becomes "cast it without paying its mana cost; the game log says so". The labels are unchanged (ADR 0111 §10).

**The server's default stays permissive.** A payload with neither flag is still waived. `gamecli`, the admin tools and the e2e helpers send bare payloads, and the setting lives on the client, where ADR 0011 §1 put it. The default flip is a client change.

**Existing players** (settings v16). `SETTINGS_VERSION` becomes 16, and the migration writes `strictMana = true` on any blob stored below 16, the v14→v15 pattern. A stored `false` cannot be told apart from the old default, so everyone moves to strict once. From v16 on the stored choice is honoured, so a player who turns it off stays off. The account copy goes through the same `migrate`, so a synced `false` written by a v15 client moves too. A stale v15 tab that writes its copy after the upgrade writes version 15, and the next v16 load moves it to strict again. That errs toward the default. The Settings help is rewritten:

> On (the default): a spell or ability costs what it says. Clicking or dragging a card taps your lands for it, spending mana already in your pool first. A card your board can't pay for is dimmed; right-click it for "Cast anyway (don't pay)", which the game log shows to the table. Off: the sandbox — mana is tracked on paper and nothing is charged.

The label stays "Strict mana enforcement".

**The practice table and the tutorial** (amends ADR 0076 §2.1 step 1 and §2.2). The practice table keeps its four forced settings and their restore on every exit path, but forces `strictMana` **on**: the tutorial teaches the table the player will meet. Step 1's body becomes: "You are seated against a practice bot. Click a card your lands can pay for and the game taps them for you. You can undo, so nothing here can go wrong." The header note at `tutorialSteps.ts:35-38` is rewritten to say the move list and the cast agree. No step changes its anchor, predicate or event. Step 5 still taps a Forest, and the pool top-up means step 6's one-drop is paid from that {G} rather than from a second land. ADR 0076's Context line "`strictMana` is off by default" becomes history.

**Bots.** Nothing changes for them. They send the enumerator's moves verbatim, and every cast move was strict and auto-tapped already (`legal/cast.go:14-16`). They gain the pool top-up through the shared affordability check (§1), and the #2163 offers through `CastOffersForLocked` (§3).

### 2. Cast anyway (don't pay)

**Where.** Whenever `strictMana` is on, every card the viewer could cast gets one more row in its right-click popover:

| Surface | Which cards |
|---|---|
| Hand (`Hand.svelte`) | the viewer's own cards with a castable face (not a card that can only be played as a land) |
| The castable-from-other-zones strip (`ExileStrip.svelte`) | its exile entries whose verb is "cast", and its commanders (#2207) |
| The command-zone panel (`CommandZone.svelte`) | the viewer's own visible commander |

A card gains a menu where it had none. `Card.svelte` takes a new optional `onCastAnyway` (and a `castAnywayBlocked` reason), and `hasMenu` counts it, so a right-click on any castable card opens the popover. The row draws no pip and does not light the ready ring.

**The row.** It sits in the popover's Sandbox section (ADR 0117 §3): an unpaid cast is a sandbox override, as a raw tap is. On a hand card it is the section's only row. The text and the accessible name are **`Cast anyway (don't pay)`**, a new name (AGENTS.md §5), as a `menuitem` with `data-cast-anyway`. Its title is "Cast it without paying its mana cost. The game log shows the table." With admin overrides on, a right-click opens the override menu instead, so the same row is added there, in a "cast" section above "move to" for a hand card. One builder serves both, as `specialActionItems` does.

**When it is greyed.** The row is offered on every such card, payable or not (owner decision 2). It is greyed, with the reason, only when something other than mana would refuse the cast. `timing.ts` gains `castAnywayBlocked(card, snap, viewerID, zone)`, which is `canCastFromHand`'s denials without its move-list shortcut and without its last "timing or mana" hint: no priority, split second, the land-only case, `cant_cast`, no legal target, no castable mode, nothing to discard or sacrifice, no payable either/or branch, and a card with no mana cost and no alternative cost (CR 118.6, the server's `ErrNoManaCost`). In the strip, an exile entry is also greyed when its `castable_here` is false, which is the server's zone, timing and gate answer without mana. Whatever the client cannot judge, a closed sorcery-speed window above all, the server refuses with its own error and the existing rejected toast.

**It asks first** (owner decision 6). Choosing the row casts nothing yet. It opens a confirmation as a request in the action dock, the place ADR 0111 gives every confirmation. The confirmation's dialog name is **`Cast <card> without paying its mana cost?`** (for example "Cast Craw Wurm without paying its mana cost?"). The question line says the same, and there are two buttons, **`Cast`** and **`Cancel`**. Cancel answers Escape. Cast has no key, for the reason the insufficient-mana request's Cast anyway has none (`targetingDock.test.ts:167`): no keystroke should reach an unpaid cast. The dialog name and the two button names are new names, and from PR 2 on they are a label contract (AGENTS.md §5, ADR 0111 §10). The e2e spec selects them, so renaming one is a breaking change. Cancel, Escape, or any other request taking the dock sends nothing. The builder is a pure function beside `insufficientManaRequest` in `targetingDock.ts`, `castAnywayConfirmRequest(cardName, handlers)`. The dock's own Cast anyway, on a refused cast, does not ask again: it is already the second step of a choice the player made.

**What it does.** After Cast, it starts the ordinary cast chain, `Board.svelte`'s `handlePlayCard`, with the card's zone and a new `CastChoices.forceCast`. The face, alternative cost, additional costs, X, modes and targets are asked exactly as for a click. `applyCastChoices` turns the flag into `strict: true, force_cast: true` on the wire, which `stampManaEnforcement` then leaves alone. No mana is spent and the pool is untouched (`mutations.go:1800-1820`). Life for Phyrexian symbols and every additional cost are still paid, because `ForceCast` waives the mana gate and nothing else. The server still judges timing, targets and every other gate: `force_cast` waives only the mana. A cast it refuses shows the existing rejected toast. The log line is the table's check on what was cast.

**The log line.** `EventCast` gains an additive field, `Unpaid bool` (`json:"unpaid,omitempty"`), set when the cast was made with `ForceCast`. A permissive cast does not set it: owner decision 2 ties the line to `force_cast`, and a permissive seat may be paying on paper. `projectEvent` copies it to `LogEvent.Unpaid` (`unpaid`, omitted when false), and `renderLogText` appends the owner's words (owner decision 4): "<player> cast <card> without paying its mana cost". For example, "Alice cast Craw Wurm without paying its mana cost", or "Alice cast Craw Wurm from exile without paying its mana cost". The dock's existing Cast anyway sends `force_cast` too, so it gets the same line. Redaction is unchanged: the flag is public, because the pool and the cost are public, and the card's name is redacted per viewer as for any cast entry (a face-down spell stays "a card"). The snapshot records `events[].unpaid` as an additive field of schema 7 (`-update-shape`, no version bump), and `docs/protocol.md` documents the wire field.

### 3. The granted-alternative-cost seam (#2163)

**The declaration.** A new catalog declaration on `Spec`, carried to `CardDef` like `CastPermissions`:

```go
// game: a static ability of a permanent that offers its controller one
// more alternative cost (CR 118.9) for each spell they cast.
type GrantedAlternativeCost struct {
	Offer AlternativeCost // Key, Label, ManaCost; Life and the card components stay unused
	Zones []ZoneKind      // the zones it reaches; nil means every zone a spell is cast from
}

var CatalogGrantedAlternativeCosts func(abilityKey string) []GrantedAlternativeCost
```

Two effects helpers in a new mechanic-named, append-only file, `effects/granted_alternative_cost.go`: `PayWUBRGForSpellsYouCast()` (key `granted-wubrg`, `ManaCost: "{W}{U}{B}{R}{G}"`, `Zones: nil`) and `CastFromHandWithoutPayingManaCost()` (key `granted-free`, `ManaCost: ""`, `Zones: [hand]`). An offer with no mana cost is free (`AlternativeCost.ManaCost`'s doc). The keys are on-disk identities, because they land on `StackItem.AltCost`. Like a token slug, they are never renamed or reused. They are namespaced so that a card's own "free" or Bringer key is never shadowed by `CastOffersForLocked`'s per-key dedupe (`seen` is set before the payability check, `cast_zones.go:271-274`).

**Derivation.** `grantedAlternativeCostsLocked(playerID, card, zone)` walks the permanents the player controls, as `standingCastPermissionsLocked` does, and reads `CatalogGrantedAlternativeCosts(catalogAbilityKeyOf(c))`. A permanent that has lost its abilities grants nothing (Humility on Jodah). Nothing is stored, so the offer lasts exactly as long as the source is on the battlefield under its controller. A land gets none: a land is not a spell. Each offer is a copy with `Granted: true`, a new field on `AlternativeCost`, which is never serialised (`cast_permission.go:649-650`). Its label is suffixed with the source's name, for example "Pay {W}{U}{B}{R}{G} rather than pay this spell's mana cost (Jodah, Archmage Eternal)". Two sources granting the same key give one offer, labelled after the first in battlefield order.

**Where it is claimable** (CR 118.9a, 601.2b). A granted offer replaces the mana cost, so it may be claimed exactly where the printed mana cost may be paid:

- `validateCastPathLocked(card, zone, nil, grant)` allows a claim of nothing, so the `nil` entry is in the list; and
- the cast's permission charges no price of its own: no flat `Cost` (cascade, discover, airbend) and no `AltCostKey` (Snapcaster's flashback, Bolas's Citadel).

That rule is the whole CR 118.9a story. A flashback card in a graveyard must pay flashback (rule 3 of `validateCastPathLocked`), so it gets no granted offer. A card Snapcaster gave flashback to, or one on top of the library under Citadel, must pay the permission's price (rule 4). A cascade or discover hit is already free and already alternative-costed. An impulse-exiled card, a Gravecrawler, a Future Sight top card, a hand card and a commander pay their printed cost, so they get the offer, with commander tax on top (CR 118.9d, `effectiveCostLocked` layers the tax after the swap). `validateCastPathLocked` asks the same question for an `alt` with `Granted` set, so the announce path and the list cannot disagree.

**The two hooks.** `CastOffersForLocked` appends the granted offers last, after the card's and the permissions' (announce precedence), through the same `add`. It drops a granted offer whose price equals one already listed: the same mana cost and no other component, the printed cost included (call 3). `resolveAlternativeCostLocked` gains the caster and tries the granted offers last, after the card's own, the disturb front face and the permission's. Because the price, the auto-tap preview, the auto-tapper (`applyAutoTapLocked` reads `effectiveCostLocked`), the view, the enumerator and the commander-return check all go through these two functions, nothing else needs a branch.

**Pricing and the auto-tapper.** `CastCostFor` pays `alt.ManaCost`. Cost increases and reductions apply to it (CR 118.9d, 601.2f), and X is fixed at 0 (CR 107.3b, `LocksXAtZero`). The auto-tapper colours {W}{U}{B}{R}{G} like any coloured cost (CR 106.1b). A card with no mana cost is castable through the offer (CR 118.6a, `castPaysPrintedCost`).

**Timing.** Unchanged. A granted offer carries no timing, so `CastTimingOpenLocked` judges the cast by the card and its permission, narrowed per claim as today (`ForClaim`). Omniscience's free sorcery waits for a main phase (owner decision 3).

**The client.** No new code. The offer reaches `card.alternative_costs` through `viewOfAlternativeCosts`, and the cost picker (`AlternativeCostModal`, opened by `afterFace` when the list is not empty) shows it with its label. The castable-from-other-zones strip prices it through `cast_prices` (#1389, #2202). The drag verdict reads the same list. With Jodah or Omniscience out, the picker opens on every cast: CR 601.2b makes the choice the caster's, and paying the printed cost can matter ("if mana was spent", converge).

**The bots.** The enumerator walks `CastOffersForLocked`, so it offers a cast for each granted offer it can afford. The heuristic's choice between equal moves is not changed here.

**Snapshots.** No new snapshot field. The offers are catalog data derived per query. `StackItem.AltCost` gains two new string values, which every resolution reader looks up with `AlternativeCostByKey` and ignores when the card does not print them, so a restored stack item is inert in a binary that has lost the static. `GrantedAlternativeCost` is reachable only from the catalog hook, not from `Game`, which `TestClosureFieldsReachableFromGame` confirms in PR 5. No corpus fixture changes.

**The four cards** (PR 6). One file each, oracle fixtures for these four IDs only:

| Card | Oracle ID | Declaration | Completeness |
|---|---|---|---|
| Fist of Suns | `6fd5e591-fae8-4128-a4d3-a848c8a8ffda` | `PayWUBRGForSpellsYouCast()` | Full |
| Jodah, Archmage Eternal | `8be4745e-36d8-430f-945e-c8a7fde9b4f6` | flying, `PayWUBRGForSpellsYouCast()` | Full |
| Leyline of Mutation | `caab67eb-65e7-4755-b116-6977e97f0844` | `PayWUBRGForSpellsYouCast()` | Caveats: "You can't begin the game with it on the battlefield from your opening hand — it has to be cast." (the other Leylines' sentence, CR 103.6a) |
| Omniscience | `730e39e6-c61d-48b5-8827-bfd952bf1be7` | `CastFromHandWithoutPayingManaCost()` | Full |

The registry row `granted-alternative-costs` closes, with a fragment in `docs/engine-seams/closed/` and `go test ./internal/roadmap/ -update`. `docs/adding-cards.md` gains a short subsection, "Alternative costs for every spell you cast (ADR 0118, #2163)", linked from AGENTS.md §7. Hunting Velociraptor's granted prowl stays on `ability-cost-modification`: it needs a spell filter (Dinosaur spells) and prowl's condition, which this struct can grow additively later.

### 4. Tests

**Go.**

- PR 2: a forced cast's `EventCast` has `Unpaid`; a permissive cast's and a paid cast's do not; the log renders "<player> cast <card> without paying its mana cost" exactly, from the hand and "from exile"; a face-down forced cast is redacted for an opponent and keeps the suffix; the shape guard records `events[].unpaid`.
- PR 3: {G} floating plus one Forest pays {1}{G} with `auto_tap`, tapping one Forest and leaving nothing floating; a pool that covers the cost taps nothing (today's shortcut); restricted pool mana (Ancient Ziggurat) is credited only where it may pay; `legal.EnumerateFor` lists the same cast; an activation and a special action top up too.
- PR 5: Jodah offers {W}{U}{B}{R}{G} from hand, from the command zone (with tax) and from impulse exile; not from a graveyard for a flashback card, not under Snapcaster, Citadel, cascade or discover; two sources give one offer; Jodah under Humility gives none; a Bringer under Fist lists its own key and not a duplicate; Sliver Queen under Fist lists no granted offer; Omniscience offers the free cast from hand only, keeps a sorcery's timing, fixes X at 0 and casts Ancestral Vision (CR 118.6a); a commander under Omniscience is not offered it; the enumerator lists the moves and the view stamps the offers; a stack item cast with `granted-wubrg` survives capture and restore.
- PR 6: one test per card through the real catalog, and the oracle fixtures.

**Vitest.**

- PR 2: the row's builder and `castAnywayBlocked`, one case per denial; the row on a hand card, a strip exile entry, a strip commander and the command-zone panel, and not on a land or with strict off; the popover opens on a plain spell's right-click; the row opens the confirmation and sends nothing; `castAnywayConfirmRequest` names the dialog "Cast <card> without paying its mana cost?", gives Cancel Escape and gives Cast no key; Cancel and Escape send nothing; Cast sends `strict: true, force_cast: true` through the cast chain, with X, modes and targets intact; the admin-overrides menu's row opens the same confirmation.
- PR 4: `stampManaEnforcement` stamps `auto_tap` with strict on and is byte-identical with it off; the v16 migration (a v15 `false` becomes true, a v16 `false` stays, a synced v15 copy moves); the request without "Auto-tap & cast" after an auto-tapped refusal; the practice table's forced value; step 1's copy. `practiceTable.test.ts`, `settings.test.ts`, `settingsSync.test.ts`, `manaEnforcement.test.ts`, `dragCastStamp.test.ts` and `targetingDock*.test.ts` are re-pinned where they assumed the old default.

**Playwright.** No existing spec needs a seed: none casts from the client. PR 2 adds `cast-anyway-2188.spec.ts`: with two Forests, a click on Grizzly Bears taps both and casts it, a right-click on Craw Wurm shows the menu item "Cast anyway (don't pay)", and choosing it opens the dock dialog named "Cast Craw Wurm without paying its mana cost?". Cancel closes it and the Wurm stays in hand. Choosing the row again, then Cast, puts the Wurm on the stack and writes "<player> cast Craw Wurm without paying its mana cost" in both players' logs. The spec selects the menu item, the dialog and the `Cast` and `Cancel` buttons by these names, which is what makes them a label contract. The nightly E2E runs on PR 2's and PR 4's branches before they merge (`gh workflow run "cmd_and_ctrl E2E" --ref <branch>`), and on `develop` after PR 4.

---

## Delivery

Each PR goes into `develop`, Sprint S59. PRs 1–4 are Issue #2188 and PRs 5–6 are Issue #2163.

| PR | What | Needs | Parallel with |
|---|---|---|---|
| 1 | **This ADR,** the S59 section in `docs/sprints.md`, and the AGENTS.md §3 ADR range line. Docs only. | — | — |
| 2 | **Cast anyway (don't pay)** (§2), server and client. `Event.Unpaid`, `LogEvent.unpaid`, the log line "<player> cast <card> without paying its mana cost", the shape guard line and `docs/protocol.md`; the row on the four surfaces and in the override menu, `castAnywayBlocked`, the dock confirmation "Cast <card> without paying its mana cost?" with Cast and Cancel (`castAnywayConfirmRequest`), `CastChoices.forceCast`; the Go and vitest in §4; the new e2e spec; the nightly E2E on the branch. Strict is still opt-in here, so the row is seen by players who turned it on. | 1 | 3, 5 |
| 3 | **Server: the auto-tapper tops up the pool** (§1). The planner and `canPayExcluding` plan what the pool is missing; the Go tests in §4. | 1 | 2, 5 |
| 4 | **Client: the default flip** (§1). `strictMana` defaults on; a clicked cast is stamped `auto_tap`; settings v16; the Settings help; the request's buttons and Cast anyway's title; the practice table's forced value; step 1's copy and the header note; the vitest in §4; the nightly E2E on the branch. | 2, 3 | 5, 6 |
| 5 | **Server: the granted-alternative-cost seam** (§3). `GrantedAlternativeCost`, the `Spec` field and catalog hook, `grantedAlternativeCostsLocked`, `AlternativeCost.Granted`, the two hooks and the `validateCastPathLocked` rule; the two effects helpers; the Go tests in §4. | 1 | 2, 3, 4 |
| 6 | **Server: the four cards** (§3), the registry row closed with its fragment, the `docs/adding-cards.md` subsection and its AGENTS.md §7 link. | 5 | 4 |

PR 4 waits for PR 2 so the escape exists before strict becomes everyone's default, and for PR 3 so that floating mana is not wasted. After PR 4: run the nightly E2E on `develop`, then close #2188 with evidence (the run, and a cmd-dev check that a fresh browser starts strict and that Cast anyway's line shows on the other seat). After PR 6: close #2163 with evidence (a cmd-dev Jodah casting a spell for {W}{U}{B}{R}{G} through the picker).

## Consequences

- A spell costs what it says at every table that has not opted out. Clicking a card the board can pay for taps the lands and casts it, in one click, as a drag already did.
- Floating mana is spent before lands are tapped, for players and bots alike, and the ready ring stops dimming a card that the pool and the lands can pay for together.
- Every castable card has a visible, logged escape when strict is on. It takes a confirmation, so it is never a misclick. An unpaid cast is never silent: the whole table reads it in the log.
- `OnPaper` payments become the exception. Converge, sunburst, adamant, Vexing Bauble's "no mana was spent" and every other ADR 0068 reader see real payments at most tables, where ADR 0068 had to note that most tables were permissive.
- Players who had strict off are moved to strict once and have to turn it off again if they want the sandbox. That is the price of a default that cannot tell "chose off" from "never chose".
- Jodah, Fist of Suns, Leyline of Mutation and Omniscience are real cards. Their offers are priced, auto-tapped, offered to bots and shown in the cost picker like any printed alternative cost.
- The cost picker opens on every cast while one of those permanents is out. That is one more click per spell, and it is the caster's choice under CR 601.2b.
- `AutoTapPreviewModal` is reachable only from a refused cast that was not auto-tapped, which no client surface now sends. It stays for this ADR; removing it is a follow-up.

## Out of scope

- **An unpaid activation.** There is no "Activate anyway". An activation that cannot be paid is refused, as it has been since #1296 for strict players.
- **Marking permissive casts in the log.** A seat with strict off still casts on paper with no line (§2, the log line).
- **A confirm before the auto-tapper spends a costly source** (a Treasure, a Spirit Guide, a once-per-turn free source). Drag casts have spent them since #1508. The planner already ranks them last (ADR 0011's amendments).
- **Hunting Velociraptor** and any granted offer with a spell filter or a condition.
- **The bot's choice between a free and a paid offer.** The arena can measure it if it matters.
- **Removing `AutoTapPreviewModal`** and the lock-tap UI inside it.
- **The opening-hand action** (CR 103.6a) for any Leyline.

## Calls made here

### Settled by the owner in review (2026-10-04)

The first draft made ten calls. The owner answered four of them in review of PR #2221 on 2026-10-04, and they are now owner decisions 4–7:

- **The log wording** (owner decision 4; was call 3). "<player> cast <card> without paying its mana cost". Only a forced cast gets the line (§2): the line belongs to `force_cast` (owner decision 2), and a permissive seat may be paying on paper. The wording is exact: only the mana is waived, and Phyrexian life and additional costs are still paid.
- **The migration** (owner decision 5; was call 4). Settings v16 moves everyone to strict once, synced copies included. A later opt-out sticks.
- **Cast anyway asks first** (owner decision 6; was call 7, which proposed no confirmation). The row opens a dock confirmation named "Cast <card> without paying its mana cost?", with Cast and Cancel. It stays in the popover's Sandbox section and in the admin override menu, and it is greyed only for reasons other than mana that the client can see. The server still judges timing.
- **The practice table forces strict on** (owner decision 7; was call 5).

### Still open

These are this ADR's calls. Each is decided above, and the owner can still overturn any of them:

1. **A clicked cast auto-taps under strict** (§1), with no preview. The owner asked for "strict payment with auto-tap"; drags and activations already work this way. The two-step "refused, then Auto-tap & cast" flow survives only for a payload without `auto_tap`.
2. **The pool tops up** (§1, PR 3). Not asked for, but strict by default makes the existing gap hit every player who taps a land first, the tutorial's step 5 included. It is a server change to the planner and the enumerator together.
3. **Duplicate offers are dropped** (§3): a granted offer priced the same as one already listed, the printed cost included (Sliver Queen under Fist), or a second source granting the same key. Two identical rows would only ask the player to choose between equals.
4. **The server's default stays permissive** (§1). The setting is the client's, and bare payloads from tools and tests keep working.
5. **The Settings label stays "Strict mana enforcement".** Only its help text changes.
6. **Granted offers are their own declaration,** not a `CastPermission` (§3). A permission is the reason a cast is legal from a zone and may carry one price, while these statics add a price wherever the printed cost may already be paid. Folding them into permissions would have put two prices on one permission and needed a standing hand permission that `storedHandPermissionLocked` deliberately does not have.
