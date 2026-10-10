# ADR 0144 — Choose a Background and the partner pairings: one rule for two commanders

**Status:** Accepted · 2026-10-09 · S58 — Deck requests, October batch
**Issues:** [#2874](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2874) (seam: Choose a Background commander pairing). Deck requests: Karlach, Fury of Avernus in Lercron-Xenagos ([#2061](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2061)) and Mishra ([#2033](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2033)). Registry row: `choose-a-background`.
**Owner decisions:** none needed. The issue left one question to this ADR, whether Friends forever and Doctor's companion belong in the same pairing rule: they do, and so does plain partner (§1). The rest is fixed by the rules.
**Numbering:** written as 0143; ADR 0143 (gameplay settings overhaul, #2888) reached `develop` first, so this ADR moved to **0144**. After `git fetch --all`, no remote branch, local branch or worktree has another ADR above 0143.
**Builds on:** the "Partner with" pairing of #2142 (`deck/partner.go`, no ADR of its own), [ADR 0093](0093-abilities-granted-to-other-permanents.md) (a layer-6 grant of a catalog bundle to another permanent, which every Background prints), [ADR 0115](0115-commanders-die.md) (a commander's trip back to the command zone), [ADR 0126](0126-bots-that-play-their-decks.md) (how the bot prices a cast).

---

## Context

Karlach, Fury of Avernus shipped with a caveat: "Choose a Background isn't supported — Karlach can be your commander, but not alongside a Background." Jaheira, Ganax and Wilson carried the same one. Five plain-partner cards carried "Partner isn't supported" or "Partner does nothing", because `deck.Resolve` refused any commander whose text said "Partner" at all. #2142 taught the deck package one partner ability, "Partner with [name]", and wrote down that the other four were still missing.

### The rules

From the pinned Comprehensive Rules (`MagicCompRules 20260925.txt`):

- **CR 702.124a:** "Each partner ability allows you to designate two legendary cards as your commander rather than one. Each partner ability has its own requirements for those two commanders. The partner abilities are: partner, partner—[text], partner with [name], choose a Background, and Doctor's companion."
- **CR 702.124c:** the deck's colour identity is the two commanders' combined identity. **CR 702.124d:** otherwise they work independently, with their own tax and their own commander damage.
- **CR 702.124f:** different partner abilities never combine. **CR 702.124g:** a card with two uses one of them, and nothing ever makes a third commander.
- **CR 702.124h:** partner pairs two cards that both have partner. **CR 702.124i:** partner—[text] pairs two cards with the same partner—[text] (Character select, Father & son, Friends forever, Survivors).
- **CR 702.124k:** "'Choose a Background' means 'You may designate two cards as your commander rather than one if one of them is this card and the other is a legendary Background enchantment card.' You can't designate two cards as your commander if one has a 'choose a Background' ability and the other is not a legendary Background enchantment card, and legendary Background enchantment cards can't be your commander unless you have also designated a commander with 'choose a Background.'"
- **CR 702.124m:** Doctor's companion pairs two legendary creature cards, one with the ability and the other "a legendary Time Lord Doctor creature card that has no other creature types". Time Lord is one creature type of two words (CR 205.3m).

Every Background prints "Commander creatures you own have …" (or "get …"). It is a static ability of the enchantment and works only while the Background is on the battlefield (CR 113.6), on each commander creature its controller owns, whoever controls that creature.

### What the engine had

- **Deck import.** `deck.Resolve` refused, as `unsupported_mechanic`, any commander whose text had the word "Partner" outside a "Partner with" line, and any companion. `deck.Validate` accepted two commanders only for a partner-with pair, and judged each commander by `isLegalCommander`, which passes any legendary card, a lone Background included.
- **The game.** Nothing assumed one commander. The command zone holds every card the deck marks `IsCommander`; any card there can be cast with its own tax (`Player.CommanderCasts` is per instance); the CR 903.9 returns, commander damage and "your commander" readers all work per commander, as partner with already showed. A noncreature commander had been castable from the command zone since planeswalker commanders.
- **The Backgrounds.** Agent of the Iron Throne and Haunted One were real ADR 0093 grants. Inspiring Leader's anthem is on the Background, gated on the right commander, and is exact. Guild Artisan's trigger lived on the Background, gated on a commander you also controlled, so a stolen commander lost it, and it carried a caveat for that.

## Decision

### 1. One pairing rule covers all five partner abilities

`deck/partner.go` reads every partner ability off the oracle text (`partnerAbilities`), one keyword line each, reminder text dropped, the way #2142 read partner with. `pairedBy(a, b)` is true when **one** ability pairs the two cards in its own terms (CR 702.124h–m): both have partner; both have the same partner—[text]; each partner with names the other; one has choose a Background and the other is a legendary Background enchantment card; or one is a legendary creature card with Doctor's companion and the other a legendary creature card whose creature types are exactly "Time Lord Doctor". Abilities never combine, so a partner card and a partner-with card are not a pair (CR 702.124f).

The issue asked whether Friends forever and Doctor's companion belong in this rule. They are the same shape, a check on two cards' printed text, and each is two or three lines here. Leaving them out would have meant keeping `Resolve`'s refusal of the word "Partner", which is what kept plain partner out too, so all three come in: refusing Friends forever while accepting Choose a Background would have been the odd choice. Nothing in the game changes for any of them; CR 702.124d's independence was already true.

`Resolve` no longer refuses a partner commander. Companion is a different mechanic (CR 702.139, a card from outside the game) and is still refused, now under its own name only.

### 2. What Validate says

- **Two commanders** go to `commanderPairViolation`. A pair `pairedBy` allows passes. Two cards with no partner ability and no Background between them are still `too_many_commanders`. Otherwise it is `invalid_partner_pair`, carrying the card at fault, the one that does not meet the other's requirement: "Karlach has Choose a Background, so its second commander has to be a Background, and X isn't one". Partner with keeps #2142's wording.
- **One commander**: a legendary Background enchantment card alone is `not_a_legal_commander` (CR 702.124k), unless it has choose a Background itself. Faceless One is both, may be your only commander, and may choose another Background. A card with choose a Background, partner or Doctor's companion is an ordinary commander alone: each ability is a permission.
- Colour identity is the union, as #2142 made it.

`Resolve` lists a Background after the commander that chose it (`backgroundsLast`, stable), whatever order the decklist gave. The command zone, the deck's fallback name ("Karlach, Fury of Avernus / Agent of the Iron Throne") and the seat's avatar, which reads the first card in the command zone, then show the creature first. Moxfield lists the commanders board alphabetically, which put the Background first.

### 3. Background grants are ADR 0093 grants

A Background's "Commander creatures you own have [ability]" is `GrantAbilities(commanderCreatureYouOwn, key)`, `grantToCommanderCreaturesYouOwn` in `effects/choose_a_background.go`: a layer-6 grant from the Background to each creature that is a commander its controller owns, under anyone's control. The granted ability is the creature's own, so "this creature" is the commander and "you" is whoever controls it now. A stolen commander keeps the ability and it works for the thief, as printed. It comes and goes with the Background on the battlefield; a Background in the command zone grants nothing. "Commander creatures you own get +N/+N and are [type]" (Raised by Giants) is two plain statics from the Background, layer 4 and layer 7b, on the same predicate.

Guild Artisan moves to this posture and loses its caveat. Inspiring Leader stays as it is: its anthem on the Background, gated on the right commander's controller, gives the printed result.

`wheneverThisAttacksAPlayerNoOpponentRicher` is the Baldur's Gate attack trigger three Backgrounds grant ("whenever this creature attacks a player, if no opponent has more life than that player"), with its intervening if read twice (CR 603.4).

### 4. Client and bot

- **Client.** No change. The deck form already said "commanders: A / B" for a pair, the command zone already holds two cards, and each commander's hover already shows its own commander damage.
- **Bot.** The heuristic already prices a cast from the command zone (ADR 0126 adds the commander bonus and the forward tax), and a Background is a permanent spell like any other. A test pins that a bot seat casts its Background from the command zone rather than passing.

### 5. Cards

- Lifted to Full: Karlach, Fury of Avernus, Jaheira, Friend of the Forest, Ganax, Astral Hunter and Wilson, Refined Grizzly (choose a Background); Anara, Wolvid Familiar, Breeches, Brazen Plunderer, Ghost of Ramirez DePietro, Malcolm, Keen-Eyed Navigator and Sakashima of a Thousand Faces (partner); Guild Artisan (§3).
- New Backgrounds, each needing nothing beyond §3: Flaming Fist, Sword Coast Sailor, Agent of the Shadow Thieves, Candlekeep Sage, Clan Crafter and Raised by Giants.

## Consequences

- Every Commander-legal partner pairing imports, and a lone Background is refused for the reason the rules give.
- `invalid_partner_pair` now covers every partner ability. The client keys nothing off it beyond showing the message.
- Follow-ups, card work rather than seams: the other Backgrounds (Veteran Soldier's tokens attacking each opponent, Popular Entertainer's goad, Criminal Past's +X/+0, Cultist of the Absolute's ward—pay life, Master Chef's entry counters, Dungeon Delver, Scion of Halaster, Shameless Charlatan and the rest) and the other choose-a-Background, partner, Friends forever and Doctor's companion cards. Wizard from Beyond's "Create a Character" is an Alchemy rule this does not read.
- Companion stays refused.
