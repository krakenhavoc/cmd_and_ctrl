package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr_b3_helpers.go — shared bodies for the Commander deck cards of
// batches 3 and 4 (#5368).

// cmdrB3OtherNontokenCreaturesYouControl is "other nontoken creatures you
// control" (Thraben Watcher): object identity for "other", the token flag
// for "nontoken".
func cmdrB3OtherNontokenCreaturesYouControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.InstanceID != source.InstanceID && target.IsCreature() &&
		target.Controller == source.Controller && !IsToken(*target)
}

// cmdrB3ZombiesYouControl counts the Zombies the effect's controller controls
// now (effective subtypes, so a changeling counts), read at resolution
// (Graveborn Muse). "Zombie" is a word of the source's text (CR 612.2).
func cmdrB3ZombiesYouControl(ctx *Context) int {
	word := "Zombie"
	if src, ok := ctx.Game.LookupCardForEffect(ctx.Source()); ok {
		word = src.TextSubtype("Zombie")
	}
	you := ctx.Controller()
	n := 0
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == you && c.IsCreature() && c.HasSubtype(word) {
			n++
		}
	}
	return n
}

// cmdrB3OtherArtifactsYouControl counts the artifacts the activator controls
// other than the ability's source (Tamiyo's Logbook).
func cmdrB3OtherArtifactsYouControl(q game.CostQuery) int {
	if q.Game == nil || q.Game.Battlefield == nil {
		return 0
	}
	n := 0
	for _, c := range q.Game.Battlefield.Cards {
		if c.Controller == q.Controller && c.IsArtifact() && c.InstanceID != q.Source.InstanceID {
			n++
		}
	}
	return n
}

// cmdrB3ReturnThisFromGraveyardWithCounter is "Return this card from your
// graveyard to the battlefield with a <kind> counter on it" (Razorlash
// Transmogrant): the card enters under the activator's control with the
// counter as it enters (CR 122.6, 614.1c), and a card that left the
// graveyard in response stays where it is (CR 400.7).
func cmdrB3ReturnThisFromGraveyardWithCounter(kind string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneGraveyard {
			return nil
		}
		return g.ReturnFromGraveyardWithCountersKeyedForEffect(item.SourceCardID, item.Controller, false,
			map[string]int{kind: 1}, game.KeyedTail{})
	}
}

// cmdrB3AnOpponentControlsNonbasicLands is the cost condition "an opponent
// controls N or more nonbasic lands": some single opponent of the
// activator does (Razorlash Transmogrant).
func cmdrB3AnOpponentControlsNonbasicLands(n int) CostPredicate {
	return func(q game.CostQuery) bool {
		if q.Game == nil || q.Game.Battlefield == nil {
			return false
		}
		count := map[uuid.UUID]int{}
		for _, c := range q.Game.Battlefield.Cards {
			if c.Controller != q.Controller && c.IsLand() && !c.HasSupertype("basic") {
				count[c.Controller]++
			}
		}
		for _, k := range count {
			if k >= n {
				return true
			}
		}
		return false
	}
}

// cmdrB3AttackersControllerLosesOne is "that creature's controller loses 1
// life" for a trigger whose Build carried the attacker's controller into
// Params.Player by value (Hissing Miasma, Marchesa's Decree): the loss lands
// on the right player even after the attacker has left, and not on a player
// who has left the game.
func cmdrB3AttackersControllerLosesOne(g *game.Game, item *game.StackItem) error {
	victim := item.Params.Player
	if p := g.PlayerByIDForEffect(victim); p == nil || p.Eliminated {
		return nil
	}
	return g.ChangePlayerLifeForEffect(item.SourceCardID, victim, -1)
}
