package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grim Hireling — Creature — Tiefling Rogue {3}{B}, 3/2:
//
//	"Whenever one or more creatures you control deal combat damage
//	 to a player, create two Treasure tokens.
//	 {B}, Sacrifice X Treasures: Target creature gets -X/-X until
//	 end of turn. Activate only as a sorcery."
//
// One of the two cards on #1213's variable-count sacrifice row, and
// the "Sacrifice X" half of it.
//
// The X is announced with the activation (CR 602.2b) and it lives
// NOWHERE in the mana cost — the Hireling prints {B}, flat, at every
// size. That is the sentence the component changed: an ability's X
// used to be readable off its cost string alone, and now the SACRIFICE
// clause can carry one too, exactly as a spell's X has been able to
// live outside the printed cost since Toxic Deluge's "pay X life".
// The generic demand is untouched, so a ten-Treasure activation still
// costs one black mana.
//
// The effect reads the same number back with ctx.X(), the way any
// other X ability does. It does not count the graveyard: by
// resolution the Treasures are gone, and the announcement is the only
// thing that knows how many there were.
//
// Two rules ride along and neither is written here:
//
//   - CR 118.3 — announcing X = 4 with three Treasures is a refused
//     announcement, not a cheap one. The engine validates the count
//     against the announced X with the same predicate the client's
//     picker and the bot enumerator read.
//   - CR 603.10a — the Treasures leave as ONE simultaneous exit, so a
//     Treasure-counting payoff sees the whole payment rather than a
//     dribble in wire order.
//
// The trigger is one per (damage step, player), not one per creature
// (CR 603.2c): three attackers connecting with one opponent make two
// Treasures, not six.
//
// No simplification.
func init() {
	const label = "{B}, Sacrifice X Treasures: Target creature gets -X/-X until end of turn."
	Register(Spec{
		OracleID:     "89738595-dafb-400a-bfba-91a53a37e717",
		Name:         "Grim Hireling",
		Completeness: CompletenessFull,
		// #810: the whole of the ability is X, so an announcement of
		// X=0 does nothing and the enumerator must not offer one.
		XMatters: true,
		Triggered: []game.TriggeredAbility{
			WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(nil,
				"Grim Hireling — create two Treasures", Do(CreateToken{
					Template: TreasureToken(),
					N:        2,
				})),
		},
		Activated: []ActivatedAbility{{
			Label:        label,
			Cost:         Plus(ManaCost("{B}"), SacrificeX("X Treasures", isTreasure)),
			SorcerySpeed: true,
			Targets:      TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				x := ctx.X()
				if x <= 0 {
					return nil
				}
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return BoostUntilEOT{
							Target:    t.ID,
							Power:     -x,
							Toughness: -x,
							Label:     "Grim Hireling — -X/-X",
						}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
}
