package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Malevolent Hermit // Benevolent Geist (#1855, ADR 0107 §4) — a
// disturb card.
//
// Front face, Creature — Human Wizard {1}{U}, 2/1:
//
//	"{U}, Sacrifice this creature: Counter target noncreature spell
//	 unless its controller pays {3}.
//	 Disturb {2}{U} (You may cast this card from your graveyard
//	 transformed for its disturb cost.)"
//
// Back face, Creature — Spirit Wizard, 2/2:
//
//	"Flying
//	 Noncreature spells you control can't be countered.
//	 If Benevolent Geist would be put into a graveyard from anywhere,
//	 exile it instead."
//
// The activation sacrifices the Hermit as a cost, so the Hermit is in
// the graveyard, ready to be disturbed, before the counter resolves.
// "Unless its controller pays {3}" is the shared CounterUnlessPaid
// prompt. The back face's static is the counter gate ADR 0106 PR 2
// built (Spec.SpellsCantBeCountered), read off the battlefield.
// Disturb is effects.Disturb; the exile clause is DisturbedExile.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         malevolentHermitOracleID,
		Name:             "Malevolent Hermit",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Disturb("{2}{U}")},
		Activated: []ActivatedAbility{{
			Label:   "{U}, Sacrifice this creature: Counter target noncreature spell unless its controller pays {3}.",
			Cost:    Plus(ManaCost("{U}"), SacrificeThis()),
			Targets: TargetSpell("target noncreature spell", Noncreature()),
			Effect:  malevolentHermitCounter,
		}},
	})
	Register(Spec{
		OracleID:        malevolentHermitOracleID + "#1",
		Name:            "Benevolent Geist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		SpellsCantBeCountered: []game.CounterShieldStatic{
			SpellsYouControlCantBeCountered("Noncreature spells you control can't be countered.", Noncreature()),
		},
		Replacements: []game.ReplacementEffect{DisturbedExile("Benevolent Geist")},
	})
}

const malevolentHermitOracleID = "51233ade-70cd-4539-9f41-5ffab761da54"

// malevolentHermitCounter is the activation's resolution: "counter
// target noncreature spell unless its controller pays {3}". A target
// that has left the stack is a fizzle the engine already handled
// (CR 608.2b), so an empty target list does nothing.
func malevolentHermitCounter(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	return CounterUnlessPaid{
		StackID:  item.Targets[0].ID,
		Cost:     "{3}",
		Question: "Malevolent Hermit — pay {3} or your spell is countered",
	}.Apply(NewContext(g, item))
}
