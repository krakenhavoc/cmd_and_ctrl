package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Drivnod, Carnage Dominus — Legendary Creature — Phyrexian Horror
// {3}{B}{B}, 8/3 (slice 296-m):
//
//	"If a creature dying causes a triggered ability of a permanent you
//	 control to trigger, that ability triggers an additional time.
//	 {B/P}{B/P}, Exile three creature cards from your graveyard: Put an
//	 indestructible counter on Drivnod, Carnage Dominus."
//
// The first line is Teysa Karlov's own clause, verbatim, and
// DoublesDying(Creature()) — the shared game.TriggerDoubler
// constructor — is exactly Teysa's TriggerDoublers entry with the
// label swapped.
//
// The activated ability is Solphim's and Tekuthal's shape one
// component over: ExileFromGraveyard(3, …) reads the CR 602.2b
// activator's own graveyard (the same AbilityCost.ExileCards
// Grim Lavamancer's "Exile two cards from your graveyard" pays), the
// {B/P}{B/P} Phyrexian mana is ManaCost's own parsed-symbol handling,
// and the counter it places is CR 122.1b's indestructible counter,
// carried by b24KeywordCounterGrant exactly as Solphim's own is.
//
// No simplification.
func init() {
	doubler := DoublesDying(Creature())
	doubler.Label = "Drivnod, Carnage Dominus"
	Register(Spec{
		OracleID:        "51780f71-bf60-4208-94ea-76fa84790fb6",
		Name:            "Drivnod, Carnage Dominus",
		Completeness:    CompletenessFull,
		TriggerDoublers: []game.TriggerDoubler{doubler},
		Static:          []game.StaticAbility{b24KeywordCounterGrant("indestructible")},
		Activated: []ActivatedAbility{{
			Label:  "{B/P}{B/P}, Exile three creature cards from your graveyard: Put an indestructible counter on Drivnod, Carnage Dominus.",
			Cost:   Plus(ManaCost("{B/P}{B/P}"), ExileFromGraveyard(3, "three creature cards", func(c game.Card) bool { return c.IsCreature() })),
			Effect: putCounterOnSourceWhileOnBattlefield("indestructible", 1),
		}},
	})
}
