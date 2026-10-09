package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Andúril, Narsil Reforged — Legendary Artifact — Equipment {2} (EDHREC
// rank 2590):
//
//	"Ascend (If you control ten or more permanents, you get the city's
//	 blessing for the rest of the game.)
//	 Whenever equipped creature attacks, put a +1/+1 counter on each
//	 creature you control. If you have the city's blessing, put two
//	 +1/+1 counters on each creature you control instead.
//	 Equip {3}"
//
// Ascend on an Equipment is the static ability of a permanent (CR
// 702.131b), so the Equipment itself earns its controller the blessing
// once they control ten permanents, equipped or not.
//
// The trigger fires for the equipped creature's attack (the attachment
// is read when the attack is declared, as Argentum Armor's is). "Instead"
// is a clause of the effect, read as the trigger RESOLVES (CR 608.2), so
// a controller who earns the blessing in response puts two counters. The
// counters go on as ONE placement per creature, two at a time, so a
// counter replacement (Hardened Scales, Doubling Season) applies once
// per creature and not once per counter, and the set is taken before
// the first counter goes on (b11PutCountersOnEachCreatureYouControl).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b7d7c131-b628-4b35-b40b-1087589ebdd0",
		Name:            "Andúril, Narsil Reforged",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordAscend},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attachedCreatureAttacked(ev, source)
			},
			Key:    "Andúril, Narsil Reforged — put +1/+1 counters on each creature you control",
			Effect: andurilCounters,
		}},
		Activated: []ActivatedAbility{
			EquipAbility("{3}"),
		},
	})
}

func andurilCounters(g *game.Game, item *game.StackItem) error {
	n := 1
	if YouHaveTheCitysBlessing(g, item.Controller) {
		n = 2
	}
	return b11PutCountersOnEachCreatureYouControl(g, item, n)
}
