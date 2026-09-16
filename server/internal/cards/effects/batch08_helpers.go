package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch08_helpers.go — the shared bodies behind the card-coverage
// roadmap's batch 08 (#301, `edhrec_rank` 902–1004). Own file per the
// #231 convention; every package-level name carries the b08 prefix
// so a parallel batch cannot collide with it.
//
// What is NOT here, because main already had it: "an opponent loses
// life" is b04OpponentLostLife, the attack / combat-damage dedup is
// OncePerBatch, "the creatures you control" snapshot is
// b04CreatureIDsControlledBy, the reveal-and-tutor body is
// b06TutorToHand, "untap all lands you control" is
// untapAllLandsControlledBy, and the Treasure is tokens.go's.

// --- token templates ---------------------------------------------

// b08BlueRedElementalToken is Resculpt's 4/4 blue and red Elemental.
func b08BlueRedElementalToken() game.Card {
	return game.Card{
		Name:      "Elemental",
		TypeLine:  "Token Creature — Elemental",
		Power:     4,
		Toughness: 4,
		Colors:    []string{"U", "R"},
	}
}

// b08RedGreenElementalToken is Omnath, Locus of Rage's 5/5 red and
// green Elemental — an Elemental, so it feeds Omnath's own dies
// trigger, which is the whole card.
func b08RedGreenElementalToken() game.Card {
	return game.Card{
		Name:      "Elemental",
		TypeLine:  "Token Creature — Elemental",
		Power:     5,
		Toughness: 5,
		Colors:    []string{"R", "G"},
	}
}

// b08BlueFishToken is Fountainport's 1/1 blue Fish.
func b08BlueFishToken() game.Card {
	return game.Card{
		Name:      "Fish",
		TypeLine:  "Token Creature — Fish",
		Power:     1,
		Toughness: 1,
		Colors:    []string{"U"},
	}
}

// --- predicates and counts ---------------------------------------

// b08EnchantmentSpellCastByYou is the enchantress condition Sythis
// and Enchantress's Presence share with Mesa Enchantress: the
// controller cast an enchantment spell. The card is read off the
// stack, where its type line is intact. Shaped as a TriggeredAbility
// AppliesTo so the two cards can name it directly.
func b08EnchantmentSpellCastByYou(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor != source.Controller {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	return ok && spell.IsEnchantment()
}

// b08IsInstantOrSorceryCard is the SearchLibrary predicate for "an
// instant or sorcery card" (Solve the Equation).
func b08IsInstantOrSorceryCard(c game.Card) bool { return c.IsInstant() || c.IsSorcery() }

// b08ControlsLegendaryCreature is Mines of Moria's "unless you
// control a legendary creature". Post-layer types and supertypes, so
// an animated legendary artifact counts exactly as CR 205.4 would
// have it.
func b08ControlsLegendaryCreature(g *game.Game, controller uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsCreature() && c.IsLegendary() {
			return true
		}
	}
	return false
}

// b08LandsWithSubtypeControlled counts the lands `controller`
// controls that carry `subtype` — Emeria's "seven or more Plains".
// Effective subtypes, so an Urborg-style type grant counts, which is
// what the printed word means.
func b08LandsWithSubtypeControlled(g *game.Game, controller uuid.UUID, subtype string) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsLand() && c.HasSubtype(subtype) {
			n++
		}
	}
	return n
}

// b08SacrificeALand is Zuran Orb's cost — "Sacrifice a land:". Built
// through the same clause the sac-outlet costs use, so the client
// opens the same picker.
func b08SacrificeALand() game.AbilityCost {
	return game.AbilityCost{SacrificeOther: sacrificeSpec("a land", Land())}
}

// b08SacrificeAToken is Fountainport's "Sacrifice a token:".
func b08SacrificeAToken() game.AbilityCost {
	return game.AbilityCost{SacrificeOther: sacrificeSpec("a token", IsTokenPredicate())}
}

// --- effect bodies -----------------------------------------------

// b08EachOpponentDraws is Cut a Deal's first half: each opponent
// draws a card, in seat order, and the count of opponents who
// actually drew comes back. "Drew a card this way" is read as "had a
// card to draw": a player whose library is empty draws nothing (and
// loses at the next check), so they do not earn the caster a card.
func b08EachOpponentDraws(g *game.Game, item *game.StackItem) (int, error) {
	ctx := NewContext(g, item)
	drew := 0
	for _, opp := range ctx.Opponents() {
		p := g.PlayerByIDForEffect(opp)
		if p == nil || p.Library == nil || p.Library.Size() == 0 {
			continue
		}
		if err := (DrawCards{Player: opp, N: 1}).Apply(ctx); err != nil {
			return drew, err
		}
		drew++
	}
	return drew, nil
}

// b08SacrificeAllMatching is "each player sacrifices all <predicate>"
// — All Is Dust. A sacrifice, not a destruction, so indestructible
// does not save anything and no "destroyed" count is needed. The set
// is snapshotted before anything moves (CR 608.2), then sacrificed
// one permanent at a time in battlefield order: the engine's
// simultaneous-exit batch is destroy / exile / bounce only, so a
// dies-watcher swept by the same spell sees only the permanents that
// left after it. Weaker than printed for that watcher's controller,
// never stronger, and declared on the card.
func b08SacrificeAllMatching(ctx *Context, match CardPredicate) error {
	for _, c := range MatchingBattlefield(ctx, match) {
		if z := ctx.Game.FindCardZoneForEffect(c.InstanceID); z == nil || z.Kind != game.ZoneBattlefield {
			continue
		}
		if err := (SacrificePermanent{Target: c.InstanceID}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// b08DoubleCountersOnEachCreatureYouControl is Bristly Bill's
// activated ability: for each creature you control, put as many
// +1/+1 counters on it as it already has. Snapshots the set and the
// counts first, so a counter doubler (Doubling Season) that fires on
// the first creature cannot change what the second receives.
func b08DoubleCountersOnEachCreatureYouControl(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	type grow struct {
		id uuid.UUID
		n  int
	}
	var todo []grow
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != item.Controller || !c.IsCreature() {
			continue
		}
		if n := c.Counters["+1/+1"]; n > 0 {
			todo = append(todo, grow{c.InstanceID, n})
		}
	}
	for _, t := range todo {
		if err := (AddCounter{Target: t.id, Kind: "+1/+1", N: t.n}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// b08KrosanVergeFetch is "search your library for a Forest card and a
// Plains card, put them onto the battlefield tapped, then shuffle" —
// two searches with two predicates, run one after the other because
// one prompt cannot carry two clauses. The Forest search leaves the
// library unshuffled so the Plains search sees the same order; the
// second shuffles once, as printed. Failing to find a Forest does not
// forfeit the Plains.
func b08KrosanVergeFetch(g *game.Game, item *game.StackItem) error {
	controller, source := item.Controller, item.SourceCardID
	return SearchLibrary{
		Player:        controller,
		Predicate:     IsLandWithSubtype("Forest"),
		Dest:          game.ZoneBattlefield,
		Limit:         1,
		Reveal:        true,
		TappedOnEntry: true,
		Source:        source,
		Reason:        "Krosan Verge — a Forest card, onto the battlefield tapped",
		Then: func(g *game.Game, _ []uuid.UUID) error {
			return g.SearchLibraryThenForEffect(game.SearchLibrarySpec{
				Player:        controller,
				Source:        source,
				Pred:          IsLandWithSubtype("Plains"),
				Dest:          game.ZoneBattlefield,
				Limit:         1,
				Reveal:        true,
				Shuffle:       true,
				TappedOnEntry: true,
				Reason:        "Krosan Verge — a Plains card, onto the battlefield tapped",
			})
		},
	}.Apply(NewContext(g, item))
}

// --- cycle shapes ------------------------------------------------

// b08OverlookLand is the Streets of New Capenna "Overlook" /
// "Courtyard" shape Riveteers Overlook established:
//
//	"When this land enters, sacrifice it. When you do, search your
//	 library for a basic <A>, <B>, or <C> card, put it onto the
//	 battlefield tapped, then shuffle and you gain 1 life."
//
// Same posture as the Overlook, with the same declared gap: the
// printed card is a trigger with a REFLEXIVE trigger inside it, two
// stack items with a response window between them, and here it is
// one item — sacrifice, search, gain. "When you do" is honoured: a
// land bounced in response cannot be sacrificed, so nothing is
// searched.
func b08OverlookLand(oracleID, name string, subtypes ...string) Spec {
	pred := func(c game.Card) bool {
		if !IsBasicLand(c) {
			return false
		}
		for _, s := range subtypes {
			if c.HasSubtype(s) {
				return true
			}
		}
		return false
	}
	reason := name + " — a basic " + subtypes[0]
	for i := 1; i < len(subtypes); i++ {
		if i == len(subtypes)-1 {
			reason += ", or " + subtypes[i]
		} else {
			reason += ", " + subtypes[i]
		}
	}
	return Spec{
		OracleID:     oracleID,
		Name:         name,
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The sacrifice and the search happen as one ability, so there is no separate chance to respond between them."},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, name+" — sacrifice it, fetch a basic tapped, gain 1 life",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneBattlefield {
							return nil
						}
						if err := (SacrificePermanent{Target: item.SourceCardID}).Apply(ctx); err != nil {
							return err
						}
						source, controller := item.SourceCardID, item.Controller
						return SearchLibrary{
							Player:        controller,
							Predicate:     pred,
							Dest:          game.ZoneBattlefield,
							Limit:         1,
							Reveal:        true,
							Shuffle:       true,
							TappedOnEntry: true,
							Reason:        reason,
							Then: func(g *game.Game, _ []uuid.UUID) error {
								return g.ChangePlayerLifeForEffect(source, controller, 1)
							},
						}.Apply(ctx)
					})
			},
		}},
	}
}

// b08FilterLand is the Shadowmoor / Eventide filter land, the shape
// Graven Cairns established:
//
//	"{T}: Add {C}.
//	 {A/B}, {T}: Add {A}{A}, {A}{B}, or {B}{B}."
//
// The hybrid cost is paid from the pool as printed (no auto-tap into
// it — CR 605.3a), and the output is two independent {A|B} picks,
// which is exactly the printed three-way choice. The colorless half
// sits at index 0 so the auto-tapper reaches for it and never spends
// floating mana on a filter.
func b08FilterLand(oracleID, name, a, b string) Spec {
	return Spec{
		OracleID:     oracleID,
		Name:         name,
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			painlessColorless(),
			{
				Cost:                    ManaAbilityCost{Tap: true, Mana: "{" + a + "/" + b + "}"},
				Produced:                "{" + a + "|" + b + "}{" + a + "|" + b + "}",
				Label:                   "{" + a + "/" + b + "}, {T}: Add {" + a + "}{" + a + "}, {" + a + "}{" + b + "}, or {" + b + "}{" + b + "}",
				IgnoreCommanderIdentity: true,
			},
		},
	}
}
