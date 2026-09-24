package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Titanic Growth — Instant for {1}{G}:
//
//	"Target creature gets +4/+4 until end of turn."
//
// Giant Growth for one more mana and one more point. It earns its
// slot in this sprint rather than looking like filler because of
// what it does to a COMMANDER: the 21-damage clock (CR 903.10a)
// makes every point of power on a commander worth roughly two, and a
// single Titanic Growth turns a 5/5 commander's four connections
// into three.
//
// Note the target clause is plain "target creature" — no "you
// control". Titanic Growth really can be cast on an opponent's
// creature, which is occasionally how you make their attacker
// survive a block you want it to survive, or how you push a
// blocked creature over a lethal threshold in a political game.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "61e09dd9-7870-48c2-9177-d6abc3162692",
		Name:         "Titanic Growth",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return BoostUntilEOT{
				Target:    item.Targets[0].ID,
				Power:     4,
				Toughness: 4,
				Label:     "Titanic Growth — +4/+4",
			}.Apply(ctx)
		},
	})
}
