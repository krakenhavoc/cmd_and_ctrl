package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// support.go — the keyword action support N (CR 701.41, #3246).
//
//	"Support N. (Put a +1/+1 counter on each of up to N other target
//	 creatures.)"  // on a permanent
//	"Support N. (Put a +1/+1 counter on each of up to N target
//	 creatures.)"  // on an instant or sorcery
//
// CR 701.41a: a permanent's support targets OTHER creatures; an
// instant's or sorcery's may target any. The Crowd Goes Wild's support X
// is the spell form with the count from X.

// SupportTargets is the target clause of support n: up to n target
// creatures, other than the source when `other` (a permanent's support).
func SupportTargets(n int, other bool) *game.TargetSpec {
	if other {
		return Another(TargetCreature("up to "+supportCountWord(n)+" other target creatures").WithCount(0, n))
	}
	return TargetCreature("up to "+supportCountWord(n)+" target creatures").WithCount(0, n)
}

// SupportEffect is the resolution of support: a +1/+1 counter on each
// target still legal (CR 608.2b), put there by the ability's controller.
func SupportEffect(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := nothingIfGone(g.AddCounterByForEffect(item.Controller, t.ID, game.CounterPlusOne, 1)); err != nil {
			return err
		}
	}
	return nil
}

// WhenThisEntersSupport is "When this permanent enters, support n."
func WhenThisEntersSupport(name string, n int, other bool) game.TriggeredAbility {
	return Targeting(WhenThisEnters(name+" — support "+supportCountWord(n), SupportEffect), SupportTargets(n, other))
}

func supportCountWord(n int) string {
	switch n {
	case 1:
		return "one"
	case 2:
		return "two"
	case 3:
		return "three"
	}
	return "N"
}
