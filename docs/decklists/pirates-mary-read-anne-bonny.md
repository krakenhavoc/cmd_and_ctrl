# Decklist support — Women of the Sea (Mary Read and Anne Bonny)

Source: [Moxfield](https://moxfield.com/decks/Yjlzz2_Zs0CruZU_vw8m-g) —
Izzet Pirate tribal, looting into Treasure into artifact payoffs.

This is a **triage of one real deck against the catalog**, used to
decide what the engine builds next. A deck is a better forcing
function than a card count: it says which gaps actually stop a game
from being played, rather than which cards happen to be easy.

## Status (refreshed 2026-10-01, #1112)

**Every card of this deck is now in the catalog.** The blocker groups
this doc used to carry (attack and end-step triggers, crew and Vehicles,
until-end-of-turn pumps, per-mode target slots, transform, discover and
descend, the modal-trigger ledger, the reveal-from-hand entry choice,
stack-item retarget, conditional block rules) have all shipped, so the
old "Blocked, by machinery needed" list is replaced below by what is
*left*: the declared gaps on the sixteen cards that ship `caveats`.
Everything else this doc names is `full`. The caveat wording is from
each card's `Caveats`, which stays the source of truth.

## Done (batch 1)

| Card | What it exercises |
|---|---|
| Reckless Fireweaver | artifact-ETB trigger, damage to each opponent |
| Ingenious Artillerist | same, at a different rate |
| Marauding Mako | discard trigger → +1/+1 counters |
| Glint-Horn Buccaneer | discard trigger → table ping (the activated half has since shipped) |
| Corsair Captain | ETB Treasure **and** a Layer 7c lord in one card |
| Impulsive Pilferer | dies → Treasure |
| Angrath's Marauders | first damage-**amount** replacement in the catalog |
| Faithless Looting | draw-then-discard in printed order |
| Mary Read and Anne Bonny | the commander: tap-to-loot + typed-discard → tapped Treasure |

Already in the catalog from earlier sprints: Sol Ring, Counterspell.

## Done (batch 2 — additional costs on cast)

| Card | What it exercises |
|---|---|
| Thrill of Possibility | the discard is a **cast cost**, so the payoffs trigger above the spell |
| Big Score | same cost, plus two Treasures for the artifact payoffs above |
| Unexpected Windfall | the same card at a different price; the deck runs both |

Read the Runes stayed blocked here, and the reason was wrong twice
over: nothing it does is a cast cost at all, and the per-card choice
is not a cost shape. It all happens on resolution, and a repetition
that offers two branches is Torment of Hailfire's option-pick chain
(#568) pointed at its own controller. **Shipped, `full`,** with
#1112.

## Done (batch 3 — impulse exile)

| Card | What it exercises |
|---|---|
| Ragavan, Nimble Pilferer | the first card played from a zone that isn't yours; Treasure + cast-only steal |
| Breeches, Brazen Plunderer | the same, with "play" instead of "cast" and the any-colour mana relaxation |

**The triage overcounted this group.** It listed five cards; only
three are impulse exile, and one of those needs machinery this batch
didn't build:

- **Malcolm, Alluring Scoundrel** is not impulse exile. It loots on
  combat damage and, at four chorus counters, lets you cast the
  *discarded* card for free — cast-from-graveyard with an alternative
  cost, which is a different mechanic (and closer to S29's
  alternative cast paths). **Shipped, `full`,** with #1112: ADR 0066's
  `game.CastPermission` names the one discarded card object at {0},
  the same per-instance grant cascade, madness and suspend use.
- **Coin of Mastery** is not impulse exile either — it's an
  enters-with-counters replacement plus a Treasure ability. Both
  halves are already expressible; it was mis-filed.
- **Breeches, Eager Pillager** is impulse exile in one of three
  modes, but its trigger is *modal with a once-per-turn-per-mode
  ledger* ("choose one that hasn't been chosen this turn"). That was
  new machinery on `TriggeredAbility`, not on exile. **Shipped,
  `full`,** with #1112 (#1751, #1754), together with Monument to
  Endurance.

## Done (batch 4 — everything that needed no engine work)

Fourteen cards written against machinery that already existed, once
batches 2 and 3 had landed. No engine, wire or client changes.

| Card | What it exercises |
|---|---|
| Captain Storm, Cosmium Raider | artifact-ETB watcher **plus** a target clause — first pairing |
| Imperial Recruiter | `SearchLibrary` with a power ceiling |
| Magmakin Artillerist | discard → damage to each opponent |
| Malcolm, Keen-Eyed Navigator | Pirate combat damage → a Treasure per damaged opponent |
| Scrounging Skyray | Marauding Mako with evasion |
| Weftstalker Ardent | Reckless Fireweaver widened to creatures, with "another" |
| Gemcutter Buccaneer | Pirate-ETB → tapped Treasure (Equipment half declared) |
| Solphim, Mayhem Dominus | the damage doubler, narrowed to noncombat and to opponents |
| Gamble | tutor + random discard |
| Windfall | every hand pitched before anyone draws |
| An Offer You Can't Refuse | the Treasures go to the **countered** spell's controller |
| Decaying Time Loop | count taken before the draw |
| Pull from Tomorrow | X draw, then a queued discard |
| Frantic Search | loot 2, untap up to three lands |

That was **31 of the 65 nonland cards** at the time of batch 4; see
Status above for where the deck stands now.

### A correction to batch 1

**Quicksmith Genius was never actually written.** It appears in this
deck and was listed as shipped; it is not in the registry. It is also
genuinely blocked, which is presumably why: "whenever an artifact you
control enters, you may discard a card. If you do, draw a card"
needs the draw to happen *after* the player has chosen what to
discard. `DiscardChoiceForEffect` queues the choice and returns, so
the draw would land first — turning a rummage into a loot, which is a
strictly better card. It wanted a continuation on the discard prompt,
the same shape `PayUnless.OnDecline` already has. **Shipped, `full`,**
with #1112, on reflexive triggers (#636, "when you do").

### Two engine findings from this batch

1. **Damage to a player from a spell or ability skipped the CR 614
   replacement pipeline entirely.** `DealDamageToPlayerForEffect`
   emitted its event and changed life directly. Combat damage and
   damage marked on creatures were both routed; this was the hole.
   No damage doubler or prevention shield could ever see a Lightning
   Bolt. Fixed here, because Angrath's Marauders is unimplementable
   without it — and it means Fog-style prevention now has a shot at
   direct damage too.
2. **A layer predicate must not call `Effective()` on its target.**
   The target's characteristics are mid-rebuild when `AppliesTo`
   runs, so Corsair Captain's Pirate check reads the printed type
   line. Cost: a creature *turned into* a Pirate isn't counted by the
   lord. That wants the layer engine to expose a partially-applied
   view, which it doesn't have.

## What is left (declared gaps)

Sixteen cards ship `caveats`: each does less than printed, never more
(#259), and says so on the public catalogue page. Grouped by the
machinery that would close them. "Untracked" means no open issue names
the gap.

**Paradigm** — Echocasting Symposium, Improvisation Capstone. The
spell resolves and goes to the graveyard; the free copies never
happen. Tracked by #1866.

**Mana-spent readers** — Coin of Mastery. The Treasure ability ships;
the +1/+1 counters for mana from artifacts spent to cast a creature
do not. Tracked by #1552.

**Partner** — Breeches, Brazen Plunderer; Malcolm, Keen-Eyed
Navigator; Ghost of Ramirez DePietro. Each is playable, but Partner
is not supported, so none can be your commander. Untracked.

**Alternative cast paths** — Ragavan (dash), Impulsive Pilferer
(encore). Decaying Time Loop's retrace shipped in #2528. The base
card works; the alternative cast is missing. Untracked.

**Batched triggers** — Ingenious Artillerist and Magmakin Artillerist.
"Whenever one or more X" fires once per X instead of once per batch,
so two artifacts entering together are two separate damage hits, and
a multi-card discard is separate 1s. Same totals; it matters only to
a per-hit damage booster. Captain Howler has the same gap: a
multi-card discard pumps by one card's worth of +2/+0. Untracked.

**Captain Howler, Sea Scourge** — also loses two things: `Ward—{2},
pay 2 life` is a composite cost (`effects.Ward` charges one component),
and the "that creature deals combat damage this turn, draw" payoff
needs a repeating, instance-scoped delayed trigger. Untracked.

**Gemcutter Buccaneer** — the Treasures-become-Equipment half is
missing. Untracked.

**Restricted counterspell** — Siren Stormtamer can counter only
spells; an ability that targets you or your creature can't be picked. Untracked.

**Kitesail Larcenist** — the enters trigger does not ship. The
granted Treasure mana ability is no longer the blocker (ADR 0093 PR 4,
#1603); what remains is the per-player "choose up to one" target
clause, whose slot count follows the seat count. Untracked.

**Ojer Axonil, Deepest Might** — the back face's return from the
graveyard as Temple of Power is not implemented, so the land half is
unreachable.

Cards that were blocked in the first pass and are `full` now:
Faithless Looting, Marauding Mako, Scrounging Skyray, Weftstalker
Ardent, Cyclonic Rift and Deflecting Swat; Smuggler's Copter, Magmatic Galleon, RMS Titanic, Jackdaw
and The Indomitable (Vehicles and crew); Captain Lannery Storm; Unstoppable
Plan; Academy Manufactor; Change of Fortune; Mishra's Command and
Prismari Command; Coercive Recruiter; Departed Deckhand; Bag of
Holding; Secluded Starforge; Glint-Horn Buccaneer; Cool but Rude;
Breeches, the Blastmaker; Hit the Mother Lode; Brass's Tunnel-Grinder;
Storm the Vault // Vault of Catlacan; Frostboil Snarl; Monument to
Endurance; Breeches, Eager Pillager; Malcolm, Alluring Scoundrel;
Read the Runes; Quicksmith Genius. Fable of the Mirror-Breaker //
Reflection of Kiki-Jiki is `full` as well: its token is now a
Goblin Shaman with the printed Treasure-making ability.

