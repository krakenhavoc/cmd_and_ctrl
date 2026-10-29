package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thorin, Company's Leader — Legendary Creature — Dwarf Warrior {4}{R},
// 4/5 (The Hobbit Eternal, HOC):
//
//	"Whenever a Dwarf you control deals combat damage to a player or
//	 battle, create two Treasure tokens. (They're artifacts with "{T},
//	 Sacrifice this token: Add one mana of any color.")
//	 {10}: Creatures you control gain double strike until end of turn."
//
// The trigger is per Dwarf, not per batch: each Dwarf's damage event makes
// its own two Treasures, as printed (Thorin himself is a Dwarf). "A player"
// is Bident of Thassa's combatDamageToPlayerBy and "battle" is Beamtown
// Beatstick's target check; the Dwarf is read through the Thorin's text
// (CR 612.2); combat damage is dealt before state-based actions run, so a
// Dwarf that dies from the same damage is still on the battlefield when
// the event is judged. The double strike is granted to the creatures you control as the
// ability resolves (CR 611.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7926b4c1-9a93-4193-ae6f-e75bbcd762ec",
		Name:         "Thorin, Company's Leader",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, thorinDwarfDealtCombatDamageToPlayerOrBattle,
				"Thorin, Company's Leader — create two Treasure tokens",
				Do(CreateToken{Template: TreasureToken(), N: 2})),
		},
		Activated: []ActivatedAbility{{
			Label: "{10}: Creatures you control gain double strike until end of turn.",
			Cost:  ManaCost("{10}"),
			// A team-wide double strike is a combat trick for the whole board.
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GrantKeywordUntilEOT{Match: And(Creature(), YouControl()), Keywords: []string{"double strike"},
					Label: "Thorin, Company's Leader — double strike"}.Apply(NewContext(g, item))
			},
		}},
	})
}

// thorinDwarfDealtCombatDamageToPlayerOrBattle is the trigger condition: a
// Dwarf the source's controller controls dealt combat damage to a player or
// to a battle.
func thorinDwarfDealtCombatDamageToPlayerOrBattle(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 {
		return false
	}
	dealer, ok := g.LookupCardForEffect(ev.Source)
	if !ok || !dealer.IsCreature() || dealer.Controller != source.Controller || !dealer.HasSubtype(source.TextSubtype("Dwarf")) {
		return false
	}
	if g.PlayerByIDForEffect(ev.Target) != nil {
		return true
	}
	target, ok := g.LookupCardForEffect(ev.Target)
	return ok && target.IsBattle()
}
