package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dionus, Elvish Archdruid — Legendary Creature — Elf Druid {3}{G},
// 3/3 (EDHREC rank 2498):
//
//	"Elves you control have "Whenever this creature becomes tapped
//	 during your turn, untap it and put a +1/+1 counter on it. This
//	 ability triggers only once each turn.""
//
// Every Elf you tap on your turn untaps and grows — a mana dork that
// taps for two and gets bigger doing it.
//
// Since ADR 0093 (PR 3) the ability is really GRANTED: a layer-6 grant
// of a trigger bundle to every Elf you control, Dionus included. Until
// then it was Dionus's own trigger watching every Elf, which got three
// things wrong that are right now:
//
//   - it is each Elf's ability, so an Elf that loses its abilities
//     (Darksteel Mutation, later timestamp) loses this one too;
//   - "only once each turn" is per Elf, read off the engine's
//     per-OBJECT trigger tally (TriggeredThisTurn) rather than a walk
//     of the event log;
//   - "your turn" and the trigger's controller are the ELF's
//     controller's.
//
// "Becomes tapped" reads two event kinds (ThisBecameTapped): the
// engine taps an attacker without an EventTapCard, so an attacking Elf
// untaps and grows and a vigilance Elf does not. The untap and the
// counter run in one resolution; an Elf that left the battlefield in
// response gets neither.
//
// One engine-wide limit, weaker and not specific to this card (ADR
// 0093 Decision 5): two instances of one granted bundle share one
// tally key, so an Elf granted this ability twice triggers only once
// each turn. Dionus is legendary, so the second instance needs a copy
// of him that escaped the legend rule (Sakashima the Impostor).
// Nothing about Dionus alone is simplified.
const (
	dionusGrant = "dionus-elvish-archdruid/untap-and-grow"
	dionusLabel = "Dionus, Elvish Archdruid — untap this creature and put a +1/+1 counter on it"
)

func init() {
	Register(Spec{
		OracleID:     "d5e5cf55-eb9e-4f01-9056-791215411b27",
		Name:         "Dionus, Elvish Archdruid",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: dionusGrant,
			Triggered: []game.TriggeredAbility{
				OnAny([]game.EventKind{game.EventTapCard, game.EventAttack},
					func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
						return ThisBecameTapped(ev, source, lki, g) &&
							IsYourTurn(g, source.Controller) &&
							!b11TriggeredThisTurn(g, source.InstanceID, dionusLabel)
					},
					dionusLabel, untapAndGrowSource),
			},
			Text: "Whenever this creature becomes tapped during your turn, untap it and put a +1/+1 counter on it. This ability triggers only once each turn.",
		}},
		Static: []game.StaticAbility{
			TribalAbilityGrant(TribeFilter{Tribes: []string{"Elf"}, YoursOnly: true}, dionusGrant),
		},
	})
}

// untapAndGrowSource is the granted ability's body: untap the creature
// that has it and put a +1/+1 counter on it, if it is still on the
// battlefield. A package-level func, so it captures nothing and
// resolves against whatever game an undo restored.
func untapAndGrowSource(g *game.Game, item *game.StackItem) error {
	host := item.SourceCardID
	if !onBattlefield(g, host) {
		return nil
	}
	ctx := NewContext(g, item)
	if err := (UntapTarget{Target: host}).Apply(ctx); err != nil {
		return err
	}
	return AddCounter{Target: host, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
}
