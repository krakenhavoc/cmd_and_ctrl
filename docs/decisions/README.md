# Architecture decision records

Each ADR is `NNNN-slug.md`, and its number appears in both the filename and the H1.
`server/internal/docsguard` fails CI on a duplicate number or a mismatch, and skips this README.

## Index

Every ADR, its title (the H1) and the first word of its **Status** line. Status words are the ADRs' own. An ADR that has shipped stays *Accepted*; the early ones that said *Implemented* or lowercase *accepted* were reworded. Read the ADR for amendments and partial supersessions. Numbers 0005, 0024, 0029, 0030 and 0190
have no file on any branch; numbers are never reused, so they stay empty.

| # | Title | Status |
|---|---|---|
| 0001 | [WebSocket library for the Go game server](0001-ws-library.md) | Accepted |
| 0002 | [Client framework](0002-client-framework.md) | Accepted |
| 0003 | [Auth and lobby architecture](0003-auth-and-lobby.md) | Accepted |
| 0004 | [Discord identity and slash-command bot (S12.5)](0004-discord-identity.md) | Accepted |
| 0006 | [Priority foundation (S13)](0006-priority-foundation.md) | Accepted |
| 0007 | [Stack foundation (S13.1)](0007-stack-foundation.md) | Accepted |
| 0008 | [Counter mechanics (S13.2)](0008-counter-mechanics.md) | Accepted |
| 0009 | [Smart priority auto-pass (S13.6)](0009-smart-priority-autopass.md) | Accepted |
| 0010 | [Card-effect catalog foundation (S14)](0010-card-effect-catalog.md) | Accepted |
| 0011 | [Mana pool, cost model, and auto-tapper (S15)](0011-mana-pool-and-auto-tapper.md) | Accepted |
| 0012 | [Continuous effects + layer system (S16)](0012-layer-system.md) | Accepted |
| 0013 | [Replacement effects engine (S17)](0013-replacement-effects.md) | Accepted |
| 0014 | [Combat keywords (S18)](0014-combat-keywords.md) | Accepted |
| 0015 | [Zone browser modal shell (S18.5)](0015-zone-browser-library-search.md) | Accepted |
| 0016 | [Broadcasting lobby mutations to connected clients](0016-lobby-state-broadcast.md) | Accepted |
| 0017 | [In-app bug reports as GitHub issues](0017-bug-report-button.md) | Accepted |
| 0018 | [Triggered abilities use the stack](0018-triggers-on-the-stack.md) | Accepted |
| 0019 | [Structured targeting](0019-structured-targeting.md) | Accepted |
| 0020 | [Activated abilities in the catalog](0020-activated-abilities.md) | Accepted |
| 0021 | [Additional costs to cast](0021-additional-costs.md) | Accepted |
| 0022 | [Impulse exile: playing cards from a zone that isn't yours](0022-impulse-exile.md) | Accepted |
| 0023 | [A `develop` branch with its own preview deployment](0023-develop-environment.md) | Accepted |
| 0025 | [Alternative costs to cast](0025-alternative-costs.md) | Accepted |
| 0026 | [Delayed triggers and return-from-exile (flicker)](0026-delayed-triggers.md) | Accepted |
| 0027 | [Attack triggers ride an event, not combat state](0027-attack-triggers.md) | Accepted |
| 0028 | [Right-click admin context menu for per-card overrides](0028-admin-context-menu.md) | Accepted |
| 0031 | [Progressive web app: installable shell, cached card art, nothing else offline](0031-progressive-web-app.md) | Accepted |
| 0032 | [Starting loyalty is printed card data, not catalog data](0032-planeswalkers.md) | Accepted |
| 0033 | [AI bot seat: legal-move enumeration, virtual seats, tiered policies](0033-ai-bot-seat.md) | Accepted |
| 0034 | [Multi-face cards: MDFC, transform, adventure](0034-multi-face-cards.md) | Accepted |
| 0035 | ["Until end of turn" continuous effects (S32)](0035-until-end-of-turn-effects.md) | Accepted |
| 0036 | [Attachments: Equipment and Auras](0036-attachments.md) | Accepted |
| 0037 | [Telling the player a card's rules aren't implemented](0037-unimplemented-card-signal.md) | Accepted |
| 0038 | [Protection-style keywords at the targeting choke point](0038-protection-style-keywords.md) | Accepted |
| 0039 | [Layer 4 is authoritative: the type predicates read `Effective()`](0039-layer-4-authoritative.md) | Accepted |
| 0040 | [The mana pipeline: restricted, derived, scaled and gated mana (S32)](0040-mana-pipeline.md) | Accepted |
| 0041 | [Persisting games across a deploy](0041-game-persistence.md) | Accepted |
| 0042 | [The public card catalogue, and declaring completeness](0042-card-catalog-page.md) | Accepted |
| 0043 | [Copy effects: layer 1 is the copiable-value baseline](0043-copy-effects.md) | Accepted |
| 0044 | [Surviving a deploy: the return path](0044-surviving-a-deploy.md) | Accepted |
| 0045 | [A restriction vocabulary for declarations and activations](0045-combat-restrictions.md) | Accepted |
| 0046 | [Layer 6 is authoritative: the catalog ability lookups read the layered result](0046-layer-6-authoritative.md) | Accepted |
| 0047 | [Keyboard shortcuts: one dispatcher, a server-gated keymap, overrides-only persistence](0047-keyboard-shortcuts.md) | Accepted |
| 0048 | [Cost modification and the free-spell family](0048-cost-modification.md) | Accepted |
| 0049 | [Architecture review of the card ↔ engine seam (September 2026)](0049-card-engine-seam-review.md) | Accepted |
| 0050 | [Discord sign-in before the invite (identity-only sessions)](0050-discord-login-identity.md) | Accepted |
| 0051 | [A persistent user database: people, their games, and their decks](0051-user-database.md) | Accepted |
| 0052 | [Bot decision harness and eval loop for the local model](0052-bot-decision-harness-and-eval.md) | Accepted |
| 0053 | [Combat damage beats: first strike and regular damage as two visible moments](0053-combat-damage-beats.md) | Accepted |
| 0054 | [Dice rolls and coin flips from effects](0054-dice-rolls-and-coin-flips.md) | Accepted |
| 0055 | [CR 732 loop breaker: the engine detects trigger loops by tally and suspends autopass](0055-loop-breaker.md) | Accepted |
| 0056 | [Infect, wither and toxic: damage results as -1/-1 counters and poison, inside the one damage tail](0056-infect-wither-toxic.md) | Accepted |
| 0057 | [Winning and losing the game by effect, and "can't lose" / "can't win"](0057-win-and-lose-by-effect.md) | Accepted |
| 0058 | ["Doesn't untap": untap-step restrictions, a next-untap-step marker, and stun counters](0058-doesnt-untap.md) | Accepted |
| 0059 | [Turn machinery: extra turns, extra phases and steps, and one turn identity](0059-turn-machinery.md) | Accepted |
| 0060 | [Leaving the game (CR 800.4): a departed player's objects leave with them](0060-leaving-the-game.md) | Accepted |
| 0061 | [Token creation and discard are replaceable events (CR 701.7b, CR 701.8, CR 614)](0061-token-creation-and-discard-are-replaceable-events.md) | Accepted |
| 0062 | [Abilities and special actions from the hand](0062-abilities-and-special-actions-from-the-hand.md) | Accepted |
| 0063 | [Durations of continuous effects, and control from spells and abilities](0063-durations-and-control.md) | Accepted |
| 0064 | [Emblems (CR 114)](0064-emblems.md) | Accepted |
| 0065 | [Modal and multi-target clauses](0065-modal-and-multi-target-clauses.md) | Accepted |
| 0066 | [Granted cast and play permissions](0066-granted-cast-and-play-permissions.md) | Accepted |
| 0067 | [Layer dependency ordering (CR 613.8) and effects that outlive their source (CR 613.6)](0067-layer-dependency-ordering.md) | Accepted |
| 0068 | [A stack item records the mana spent to cast it: converge, sunburst, adamant, and "if no mana was spent"](0068-the-mana-spent-on-a-spell.md) | Accepted |
| 0069 | [Face-down objects (CR 406.3a, CR 708)](0069-face-down-objects.md) | Accepted |
| 0070 | [Untap-step choices (CR 502.3)](0070-untap-step-choices.md) | Accepted |
| 0071 | [Designations that switch abilities on: Class levels, solved Cases, station thresholds, unlocked doors](0071-designations-that-switch-abilities-on.md) | Accepted |
| 0072 | [Protection (CR 702.16)](0072-protection.md) | Accepted |
| 0073 | [Optional additional costs and the announce-time cast gate](0073-optional-additional-costs-and-the-cast-gate.md) | Accepted |
| 0074 | [Triggered mana abilities: a trigger that never reaches the stack, and the colour it can read](0074-triggered-mana-abilities.md) | Accepted |
| 0075 | [Table settings and host controls: who runs a table, what they can change, and spawning on a live game](0075-table-settings-and-host-controls.md) | Accepted |
| 0076 | [The tutorial: a scripted practice game that teaches the client, not the rules](0076-tutorial.md) | Accepted |
| 0077 | [Opponent boards are read-outs, not small cards: the summary panel and what expands it](0077-opponent-board-summary.md) | Accepted |
| 0078 | [Tokens get real artwork: a Scryfall token printing, resolved at runtime](0078-token-art.md) | Accepted |
| 0079 | [Transforming a permanent: turning a card over is not a zone change (CR 701.27, CR 712)](0079-transforming-a-permanent.md) | Accepted |
| 0080 | [Attack taxes: "creatures can't attack you unless their controller pays {N}" (CR 508.1h)](0080-attack-taxes.md) | Accepted |
| 0081 | [Earthbend: a counted keyword action that animates a target, and a delayed trigger keyed to an object](0081-earthbend-and-object-keyed-delayed-triggers.md) | Accepted |
| 0082 | [Casting face down, and turning face up (CR 708.4, CR 116.2g)](0082-casting-face-down-and-turning-face-up.md) | Accepted |
| 0083 | [Triggered and static abilities on a token that is not a copy](0083-token-abilities.md) | Accepted |
| 0084 | [Phasing: a permanent that is treated as though it does not exist (CR 702.26)](0084-phasing.md) | Accepted |
| 0085 | ["Your life total can't change": a player-scoped replacement with a CR 611.2 duration](0085-life-total-cant-change.md) | Accepted |
| 0086 | [Storm, and the turn's cast order](0086-storm-and-the-turns-cast-order.md) | Accepted |
| 0087 | [Amass: find-or-create an Army, and a keyword that carries a subtype](0087-amass.md) | Accepted |
| 0088 | [Ordered library placement: one prompt for "put them on top / on the bottom in any order"](0088-ordered-library-placement.md) | Accepted |
| 0089 | [Gift: an optional cost paid by choosing an opponent](0089-gift.md) | Accepted |
| 0090 | [Preparation cards: the prepared designation and casting the prepare spell as a copy](0090-preparation-cards.md) | Accepted |
| 0091 | [Hideaway: a face-down exile its source's controller may look at, linked to that source, with a free play](0091-hideaway.md) | Accepted |
| 0092 | [A public roadmap from a CI-checked registry, and a site portal](0092-public-roadmap-and-site-portal.md) | Accepted |
| 0093 | [Abilities granted to other permanents: a layer-6 grant writes a catalog bundle key onto the recipient](0093-abilities-granted-to-other-permanents.md) | Accepted |
| 0094 | [Card lookups go through a checked hint table, not a walk of every zone](0094-card-location-index.md) | Accepted |
| 0095 | [Deck coverage checks and deck requests, on the site and in Discord](0095-deck-coverage-and-deck-requests.md) | Accepted |
| 0096 | [The monarch from a card effect](0096-the-monarch-from-a-card-effect.md) | Accepted |
| 0097 | [Modes "that hasn't been chosen"](0097-modes-that-havent-been-chosen.md) | Accepted |
| 0098 | [Discarding a card as a permanent would enter (Mox Diamond)](0098-discard-as-a-permanent-would-enter.md) | Accepted |
| 0099 | [Discover](0099-discover.md) | Accepted |
| 0100 | [Delve, either/or additional costs, and a variable sacrifice count on a cast](0100-delve-either-or-and-variable-sacrifice-costs.md) | Accepted |
| 0101 | [Keyword counters read by the engine](0101-keyword-counters.md) | Accepted |
| 0102 | [A permanent that enters under another player's control](0102-entering-under-another-players-control.md) | Accepted |
| 0103 | [Rooms: casting either half, locked doors, and unlocking them](0103-rooms.md) | Accepted |
| 0104 | [Gaining control of a spell on the stack](0104-gaining-control-of-a-spell.md) | Accepted |
| 0105 | [Highlighting every legal action](0105-legal-action-highlights.md) | Accepted |
| 0106 | [Five small seams from the S50 re-checks](0106-five-small-seams-from-the-s50-rechecks.md) | Accepted |
| 0107 | [State triggers, rebound, disturb and damage prevention](0107-state-triggers-rebound-disturb-and-damage-prevention.md) | Accepted |
| 0108 | [Turn-scoped effects, object history and the rest of the damage shields](0108-turn-scoped-effects-object-history-and-damage-shields.md) | Accepted |
| 0109 | [Rule gates, land types, mana and cost components](0109-rule-gates-land-types-mana-and-cost-components.md) | Accepted |
| 0110 | [Remember me: durable sign-in, account settings, admins and saved setups](0110-remember-me.md) | Accepted |
| 0111 | [Table UX: one action dock, bottom right](0111-action-dock.md) | Accepted |
| 0112 | [Signed-in home, player mode, and one decks page](0112-signed-in-home-player-mode-and-one-decks-page.md) | Accepted |
| 0113 | [Small seams for the S58 deck requests](0113-small-seams-for-the-s58-deck-requests.md) | Accepted |
| 0114 | [The Ring tempts you](0114-the-ring-tempts-you.md) | Accepted |
| 0115 | [Commanders die (CR 903.9a)](0115-commanders-die.md) | Accepted |
| 0116 | [A card filter on the revealed-hand pick](0116-a-card-filter-on-the-revealed-hand-pick.md) | Accepted |
| 0117 | [Click to act, and a per-colour mana stepper](0117-click-to-act-and-a-per-colour-mana-stepper.md) | Accepted |
| 0118 | [Strict payment by default, Cast anyway, and alternative costs for every spell](0118-strict-payment-by-default-and-alternative-costs-for-every-spell.md) | Accepted |
| 0119 | [A stack you can follow](0119-a-stack-you-can-follow.md) | Accepted |
| 0120 | [Expand a player's board on top of the table](0120-expand-a-players-board.md) | Accepted |
| 0121 | [Animated dice: an opening roll at the table, dice you can watch, and a Roll a die action](0121-animated-dice.md) | Accepted |
| 0122 | [An agent at the table: a local MCP seat](0122-an-agent-at-the-table-a-local-mcp-seat.md) | Accepted |
| 0123 | [Monitoring: metrics, logs, dashboards and alerts](0123-monitoring-metrics-logs-dashboards-and-alerts.md) | Accepted |
| 0124 | [Admin views: accounts, games and who is on now](0124-admin-views-accounts-games-and-who-is-on-now.md) | Accepted |
| 0125 | [A walkthrough that keeps up: first-use hints and a refreshed tutorial](0125-a-walkthrough-that-keeps-up.md) | Accepted |
| 0126 | [Bots that play their decks: the heuristic prices what a card does](0126-bots-that-play-their-decks.md) | Accepted |
| 0127 | [Answering repeated prompts for you](0127-answering-repeated-prompts-for-you.md) | Accepted |
| 0128 | [Playmats: an image on your account that shows behind your battlefield](0128-playmats.md) | Accepted |
| 0129 | [Energy: getting {E} and paying it](0129-energy-getting-and-paying-it.md) | Accepted |
| 0130 | [Exert: an optional cost to attack, a skipped untap, and the triggers that watch it](0130-exert.md) | Accepted |
| 0131 | [K'rrik: paying 2 life for {B} in any cost](0131-krrik-pay-life-for-black-mana.md) | Accepted |
| 0132 | [Day and night, daybound and nightbound](0132-day-and-night.md) | Accepted |
| 0133 | [Opening-hand actions: beginning the game with a card on the battlefield](0133-opening-hand-actions.md) | Accepted |
| 0134 | [Combat you can watch: attackers lunge, hits land](0134-combat-you-can-watch.md) | Accepted |
| 0135 | [Alternative costs that tap, discard, awaken and emerge](0135-alternative-costs-that-tap-discard-awaken-and-emerge.md) | Accepted |
| 0136 | [Planning the turn's mana: rock first, draw first, two spells over one](0136-planning-the-turns-mana.md) | Accepted |
| 0137 | [Craft: materials from two zones, the return from exile transformed, and the link to what was exiled](0137-craft.md) | Accepted |
| 0138 | [Speed: start your engines! and max speed](0138-speed.md) | Accepted |
| 0139 | [Empower Jace and the Jace planeswalker token](0139-empower-jace.md) | Accepted |
| 0140 | [Eminence: a static that works from the command zone](0140-eminence-statics-from-the-command-zone.md) | Accepted |
| 0141 | [Bestow: a creature cast as an Aura](0141-bestow.md) | Accepted |
| 0142 | [Declared answers on catalog abilities](0142-declared-answers-on-catalog-abilities.md) | Accepted |
| 0143 | [Gameplay settings overhaul](0143-gameplay-settings-overhaul.md) | Accepted |
| 0144 | [Choose a Background and the partner pairings: one rule for two commanders](0144-choose-a-background-and-the-partner-pairings.md) | Accepted |
| 0145 | [Meld: one permanent, two cards (CR 701.42, CR 712.4)](0145-meld.md) | Accepted |
| 0146 | [Voting: will of the council and council's dilemma (CR 701.38)](0146-voting.md) | Accepted |
| 0147 | [Auto-assign combat damage, and assigning it on the blockers](0147-auto-assign-combat-damage.md) | Accepted |
| 0148 | [Capacity limits for a public launch: a weighted table-load cap, a connection cap, creation rates, and a restore pass that respects them](0148-capacity-limits-for-a-public-launch.md) | Accepted |
| 0149 | [Privacy notice and self-serve account deletion: a tombstone, not a delete; anonymised past games; a notice that cannot drift](0149-privacy-notice-and-account-deletion.md) | Accepted |
| 0150 | [Other players pay: "unless any player pays" and assist](0150-other-players-pay-any-player-offers-and-assist.md) | Accepted |
| 0151 | [Flip cards: a status, not a face (CR 710, CR 110.5)](0151-flip-cards.md) | Accepted |
| 0152 | [More Than Meets the Eye, convert and living metal (CR 702.162, 701.28, 702.161)](0152-more-than-meets-the-eye-convert-and-living-metal.md) | Accepted |
| 0153 | [A discard an opponent causes: who controls the cause, and a discard that enters the battlefield (CR 614.1a, CR 701.9a)](0153-a-discard-an-opponent-causes.md) | Accepted |
| 0154 | [Auras that enchant a card in a graveyard (Animate Dead)](0154-auras-that-enchant-a-card-in-a-graveyard.md) | Accepted |
| 0155 | [Having another object's activated abilities: a layer-6 borrow that names the lent row](0155-having-another-objects-activated-abilities.md) | Accepted |
| 0156 | [A front door: one site shell, a front page, the C2 Discord, and sign-up gated by a Discord role](0156-a-front-door-site-shell-front-page-and-a-discord-gate.md) | Accepted |
| 0157 | [Crimes at announcement](0157-crimes-at-announcement.md) | Accepted |
| 0158 | [Signed-in pages: Tables with History, a deck grid, and a catalog with a filter sidebar](0158-signed-in-pages-tables-with-history-decks-and-the-catalog.md) | Accepted |
| 0159 | [Per-table Discord voice rooms](0159-per-table-discord-voice-rooms.md) | Accepted |
| 0160 | [Text-changing effects (CR 612): exchanging text boxes is a layer-3 substitution of the ability key](0160-text-changing-effects-exchanging-text-boxes.md) | Accepted |
| 0161 | [An LFG board in Discord](0161-an-lfg-board-in-discord.md) | Accepted |
| 0162 | [Discord DM notifications](0162-discord-dm-notifications.md) | Accepted |
| 0163 | [Discord Linked Roles: CMD & CTRL stats on a member's profile](0163-discord-linked-roles.md) | Accepted |
| 0164 | [Rematch: the host offers the same game again](0164-rematch-the-host-offers-the-same-game-again.md) | Accepted |
| 0165 | [Game statistics: a record derived from every finished game](0165-game-statistics-a-record-derived-from-every-finished-game.md) | Accepted |
| 0166 | [Controlling another player: Word of Command](0166-controlling-another-player-word-of-command.md) | Accepted |
| 0167 | [Controlling another player's turn: Emrakul, the Promised End](0167-controlling-another-players-turn-emrakul-the-promised-end.md) | Accepted |
| 0168 | [The London mulligan: cards to the bottom after keep, and the free first mulligan](0168-the-london-mulligan.md) | Accepted |
| 0169 | [The life that damage costs runs the life-loss window](0169-the-life-that-damage-costs-runs-the-life-window.md) | Accepted |
| 0170 | [Choosing other permanents as a permanent enters](0170-choosing-other-permanents-as-a-permanent-enters.md) | Accepted |
| 0171 | [One mana and life ward payment](0171-one-mana-and-life-ward-payment.md) | Accepted |
| 0172 | [Companion: a card outside the game](0172-companion-a-card-outside-the-game.md) | Accepted |
| 0173 | [Shield counters](0173-shield-counters.md) | Accepted |
| 0174 | [A game results feed in Discord](0174-a-game-results-feed-in-discord.md) | Accepted |
| 0175 | [Players choose before simultaneous sacrifices](0175-players-choose-before-simultaneous-sacrifices.md) | Accepted |
| 0176 | [A movable stack panel](0176-a-movable-stack-panel.md) | Accepted |
| 0177 | [Away players: showing who is gone, and moving the table on without them](0177-away-players-and-moving-the-table-on.md) | Accepted |
| 0178 | [Life instructions after a replaced move](0178-life-instructions-after-a-replaced-move.md) | Accepted |
| 0179 | [Shop links for the real cards](0179-shop-links-for-real-cards.md) | Accepted |
| 0180 | [Secret council: ballots hidden until every player has voted](0180-secret-council.md) | Accepted |
| 0181 | [Deploy credentials: deploy runners, environment keys, and failing closed](0181-deploy-credentials-fail-closed.md) | Accepted |
| 0182 | [Your own model key for AI seats](0182-your-own-model-key-for-ai-seats.md) | Accepted |
| 0183 | [Compleated: entry loyalty from Phyrexian life payments](0183-compleated-entry-loyalty.md) | Accepted |
| 0184 | [Weighted model budget and a share per table](0184-weighted-model-budget-and-a-share-per-table.md) | Accepted |
| 0185 | [Double-faced tokens and incubate (CR 701.53, CR 111.10i)](0185-double-faced-tokens-and-incubate.md) | Accepted |
| 0186 | [Crewing and saddling: who paid, and what they count for](0186-crewing-and-saddling-who-paid-and-what-they-count-for.md) | Accepted |
| 0187 | [Unless any player pays: one offer per land, graveyard exile, and an either payment](0187-any-player-pays-per-object-graveyard-and-either-payments.md) | Accepted |
| 0188 | [Turn machinery, the rest: skipped turns, effects bound to one turn or one combat, and showing the turn](0188-skipped-turns-and-turn-bound-effects.md) | Accepted |
| 0189 | [The rest of face-down cards and manifest dread: manifesting from a hand, both ways back up, survival, and a graveyard-exile cast cost](0189-the-rest-of-face-down-and-manifest-dread.md) | Accepted |
| 0191 | [A sideboard is the outside-the-game pool](0191-a-sideboard-is-the-outside-the-game-pool.md) | Accepted |
| 0192 | [A deck names its printing, and the default avoids gimmick reprints](0192-a-deck-names-its-printing.md) | Accepted |
| 0193 | [Demonstrate a loop: record one iteration, propose N, replay it (CR 732.2)](0193-demonstrate-a-loop-record-once-replay-n-times.md) | Accepted |
| 0194 | [The undo ring: snapshot only what someone can undo](0194-snapshot-only-what-someone-can-undo.md) | Accepted |
| 0195 | [Deploys that hand each table over at a clean point: a short two-process overlap, not hours of blue/green](0195-deploys-that-hand-each-table-over-at-a-clean-point.md) | Accepted |
| 0196 | [Bug reports mirrored to a Discord forum thread](0196-bug-reports-mirrored-to-a-discord-forum.md) | Accepted |
| 0197 | [Optional triggers ask "you may" as they resolve](0197-optional-triggers-ask-as-they-resolve.md) | Accepted |
| 0198 | [Outside the game is the card database when a deck has no sideboard](0198-outside-the-game-is-the-card-database-without-a-sideboard.md) | Accepted |
| 0199 | [A hand played revealed](0199-a-hand-played-revealed.md) | Accepted |
| 0200 | [Activating a mana ability of each land](0200-activating-a-mana-ability-of-each-land.md) | Accepted |
| 0201 | [Restarting the game](0201-restarting-the-game.md) | Accepted |
| 0202 | [Splice](0202-splice.md) | Accepted |
| 0203 | [The initiative and dungeons](0203-the-initiative-and-dungeons.md) | Accepted |
| 0204 | [Seats at the table](0204-seats-at-the-table.md) | Accepted |
| 0205 | [View frames on the wire: compress before we diff, and library cards in the dump](0205-view-frame-deltas-and-library-cards-in-the-dump.md) | Accepted |
| 0206 | [Attacking only the nearest opponent in a direction](0206-attacking-only-the-nearest-opponent-in-a-direction.md) | Accepted |
| 0207 | [Attractions](0207-attractions.md) | Accepted |
| 0208 | [Stickers and ticket counters](0208-stickers-and-ticket-counters.md) | Accepted |
| 0209 | [Table effects: sound categories, effect cues and signature animations](0209-table-effects-sound-and-signature-animations.md) | Accepted |
| 0210 | [Putting a card onto the battlefield as a different type](0210-putting-a-card-onto-the-battlefield-as-a-different-type.md) | Accepted |
| 0211 | [Closing idle tables: unstarted tables after a day, abandoned games after two](0211-closing-idle-tables.md) | Accepted |
| 0212 | [A new table's lobby card as a character select](0212-a-new-tables-lobby-card-as-a-character-select.md) | Accepted |
| 0213 | [Friends, recently played with, and live presence](0213-friends-recently-played-with-and-live-presence.md) | Accepted |
| 0214 | [CMD & CTRL as a Discord Activity: the table inside the voice channel](0214-cmd-and-ctrl-as-a-discord-activity.md) | Accepted |
| 0215 | [Scheduled games: Discord Scheduled Events with RSVP seats](0215-scheduled-games-discord-scheduled-events-and-rsvp-seats.md) | Accepted |
| 0216 | [Planechase as a table option (CR 901)](0216-planechase-as-a-table-option.md) | Accepted |
| 0217 | [Advanced player stats, per deck and overall](0217-advanced-player-stats-per-deck-and-overall.md) | Accepted |
| 0218 | [A spell copy that gains a triggered ability](0218-a-spell-copy-that-gains-a-triggered-ability.md) | Accepted |
| 0219 | [An ability a resolved effect gives an exiled card](0219-an-ability-a-resolved-effect-gives-an-exiled-card.md) | Accepted |
| 0220 | [Pods: groups of players for shared stats](0220-pods-groups-of-players-for-shared-stats.md) | Accepted |
| 0221 | [Block taxes: "creatures can't block unless their controller pays {N}" (CR 509.1d)](0221-block-taxes.md) | Accepted |
| 0222 | [Formats beyond Commander: Standard 1v1 and booster draft](0222-formats-beyond-commander-standard-1v1-and-booster-draft.md) | Accepted |
| 0223 | [Mutate: merged permanents, over or under (CR 702.140, CR 730)](0223-mutate-merged-permanents.md) | Accepted |
| 0224 | [Banding: attacking in a band, blocking a band, and who divides the damage (CR 702.22)](0224-banding.md) | Accepted |
| 0225 | [Splitting the card catalog package: a core, card shards and test packages](0225-splitting-the-card-catalog-package.md) | Accepted |
| 0226 | [Subgames (Shahrazad) and dexterity cards (Chaos Orb, Falling Star)](0226-subgames-and-dexterity-cards.md) | Proposed |
| 0227 | [Conspiracies in the command zone, and hidden agenda](0227-conspiracies-in-the-command-zone-and-hidden-agenda.md) | Accepted |

## Picking a number

**Check every branch, not just the one you are on.** ADR files live in `docs/decisions/` and the
number is in the filename *and* the H1, so two branches that both grab "the next number" collide
silently. They do **not** conflict at merge time: the filenames differ, so git merges both cleanly
and neither author learns. The collision is invisible in the diff and invisible in the merge, and
only shows up when somebody lists the directory — by which point the number is in commit messages,
issue bodies and cross-links in other ADRs.

`TestADRNumbersAreUniqueAndMatchTheirHeading` (`server/internal/docsguard`) fails CI on a duplicate
number and on an H1 that disagrees with its filename. It exists because asking authors to check was
tried first and three collisions reached `main` anyway. If it fires, renumber the ADR with fewer
inbound links (`git grep` its filename), fix its H1, and update the ADR range line in [AGENTS.md §4](../../AGENTS.md#picking-an-adr-number).

```bash
git fetch --all --prune
for b in $(git branch -r --format='%(refname:short)' | grep -v HEAD); do
  git ls-tree --name-only "$b" docs/decisions/
done | sed 's|.*/||' | cut -d- -f1 | sort -u
```

Take the first number that does not appear, and say in the PR body which branches you checked.
Numbers are **not** reused when an ADR is renumbered or abandoned — `0005`, `0024`, `0029` and
`0030` are permanently unused for exactly that reason, and `0190` was skipped by ADRs 0193 to 0199
as a possible claim and never used. If a number you already used turns out to be
taken, renumber **your** file (title, filename and every inbound link) rather than asking the other
branch to move; the one that merges first keeps the number.
