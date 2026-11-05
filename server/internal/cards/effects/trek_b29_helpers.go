package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b29_helpers.go — shared pieces for Star Trek batch 29 (#4973).

// trekB29SurveilLand registers one of Star Trek's monocoloured "surveil
// lands":
//
//	"This land enters tapped unless you control two or more other lands.
//	 When this land enters untapped, choose one —
//	 • <mode>
//	 • Surveil 1.
//	 {T}: Add {<colour>}."
//
// The enters-tapped clause is the slowland condition, read as the land
// enters (EntersTappedUnless). The trigger is a real modal ETB trigger
// that fires only when the land is untapped as it sits on the
// battlefield after the replacement decided it; an untapped entry by
// another effect's say-so (a "lands you control enter untapped" static)
// triggers it too, as printed.
func trekB29SurveilLand(oracleID, name, colour string, mode game.ModeOption) {
	trigger := On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return b27SelfEnteredUntapped(ev, source)
	}, name+" — choose one", func(*game.Game, *game.StackItem) error { return nil })
	trigger.Modes = ChooseOne(mode, ModeSteps("Surveil 1.", nil, Surveil{N: 1}))
	Register(Spec{
		OracleID:     oracleID,
		Name:         name,
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{EntersTappedUnless(otherLandsAtLeast(2))},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{" + colour + "}",
			Label:    "Add {" + colour + "}",
		}},
		Triggered: []game.TriggeredAbility{trigger},
	})
}

// trekB29TargetCreatureUntilEOT is a surveil land's targeted bullet:
// "Target creature gets +p/+t and gains <keyword> until end of turn". A
// target gone illegal by resolution gets nothing (CR 608.2b).
func trekB29TargetCreatureUntilEOT(label string, power, toughness int, keyword string) game.ModeOption {
	return ModeDoing(label, TargetCreature("target creature"),
		func(_ *game.StackItem, ctx *Context, occ int) error {
			t, ok := ModeTarget(ctx, occ)
			if !ok {
				return nil
			}
			if err := (BoostUntilEOT{Target: t.ID, Power: power, Toughness: toughness, Label: label}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{keyword}, Label: label}.Apply(ctx)
		})
}

// trekB29NewPlanet is an anomaly land's last sentence, "Create a tapped
// Planet land token named New Planet with "{T}: Add one mana of any
// color."" — Genesis Effect's printed token (ttrk 27).
func trekB29NewPlanet(ctx *Context) error {
	return CreateTokenAdvanced{Spec: Token(trekB16NewPlanetToken()).EntersTapped(), N: 1}.Apply(ctx)
}

// trekB29AnomalyCost is an anomaly land's "{cost}, {T}, Sacrifice this
// land" — the sacrifice is a cost, so the land is gone before the
// ability resolves and "it" is read by last known information (CR
// 608.2h).
func trekB29AnomalyCost(mana string) game.AbilityCost {
	return Plus(ManaCost(mana), TapCost(), SacrificeThis())
}

// trekCreatureTypesAmongNonBorg counts "creature types among non-Borg
// creatures <controller> controls": each creature type once however many
// creatures share it (CR 205.3m), read through the layers. A changeling
// has every creature type (CR 702.73a), so it is a Borg and counts
// nothing. `borg` is the word as the source's text reads it (CR 612.2).
func trekCreatureTypesAmongNonBorg(g *game.Game, controller uuid.UUID, borg string) int {
	seen := map[string]bool{}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != controller || !c.IsCreature() || c.HasSubtype(borg) {
			continue
		}
		for _, st := range c.Effective().Subtypes {
			if canon, ok := game.CanonicalCreatureType(st); ok {
				seen[canon] = true
			}
		}
	}
	return len(seen)
}

// trekB29StarfleetAcademyAmount is Starfleet Academy's amount, read off
// the land's own text for the word "Borg".
func trekB29StarfleetAcademyAmount(g *game.Game, controller, source uuid.UUID) int {
	borg := "Borg"
	if c, ok := g.LookupCardForEffect(source); ok {
		borg = c.TextSubtype(borg)
	}
	return trekCreatureTypesAmongNonBorg(g, controller, borg)
}
