package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Preacher of the Schism — Creature — Vampire Cleric {2}{B}, 2/4:
//
//	"Deathtouch
//	 Whenever this creature attacks the player with the most life or
//	 tied for most life, create a 1/1 white Vampire creature token with
//	 lifelink.
//	 Whenever this creature attacks while you have the most life or are
//	 tied for most life, you draw a card and you lose 1 life."
//
// Two separate triggers, and both can fire on one attack: a Preacher
// that is ahead of the table attacking the other leader makes a token
// AND draws. "The player with the most life" is read across every player
// still in the game, the Preacher's controller included, so a controller
// who leads attacks "the player with the most life" only by tying. Both
// conditions are read once, as the attack is declared (they are trigger
// conditions, not intervening ifs), and do not re-check at resolution.
//
// The first trigger needs the attack to be at a PLAYER: an attack on a
// planeswalker or a battle is not "attacks the player". The second asks
// nothing of the target, so it fires whatever the Preacher attacks.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "74947e47-c4ff-4f47-9845-befa7403cbde",
		Name:            "Preacher of the Schism",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, preacherAttackedTheLifeLeader,
				"Preacher of the Schism — create a 1/1 white Vampire creature token with lifelink",
				Do(CreateToken{Template: b36WhiteVampireLifelinkToken(), N: 1})),
			On(game.EventAttack, preacherAttackedWhileYouLead,
				"Preacher of the Schism — you draw a card and you lose 1 life",
				b36DrawAndLoseOne),
		},
	})
}

// playerHasTheMostLife is "has the most life or is tied for most life"
// among the players still in the game.
func playerHasTheMostLife(g *game.Game, id uuid.UUID) bool {
	var mine int
	found := false
	for _, p := range g.Seats {
		if p != nil && !p.Eliminated && p.ID == id {
			mine, found = p.Life, true
		}
	}
	if !found {
		return false
	}
	for _, p := range g.Seats {
		if p != nil && !p.Eliminated && p.Life > mine {
			return false
		}
	}
	return true
}

// preacherAttackedTheLifeLeader: this creature was declared attacking a
// player who has the most life or is tied for it.
func preacherAttackedTheLifeLeader(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return attackDeclared(ev, source) &&
		g.ClassifyAttackTargetForEffect(ev.Target) == game.AttackTargetPlayer &&
		playerHasTheMostLife(g, ev.Target)
}

// preacherAttackedWhileYouLead: this creature was declared attacking
// while its controller has the most life or is tied for it.
func preacherAttackedWhileYouLead(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return attackDeclared(ev, source) && playerHasTheMostLife(g, source.Controller)
}
