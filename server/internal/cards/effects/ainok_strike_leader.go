package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ainok Strike Leader — Creature — Dog Warrior {1}{W}, 2/2:
//
//	"Whenever you attack with this creature and/or your commander, for
//	 each opponent, create a 1/1 red Goblin creature token that's tapped
//	 and attacking that player.
//	 Sacrifice this creature: Creature tokens you control gain
//	 indestructible until end of turn."
//
// "And/or" is one trigger per attack declaration however many of the
// two attack (OncePerBatch): the batch fires on the first attack event
// naming this creature or a commander its controller owns. The Goblins
// were never declared as attackers, so they trigger nothing that reads
// "whenever a creature attacks" (CR 508.4), and each attacks the
// opponent it was made for, as Adeline's Humans do.
//
// The sacrifice grant is snapshotted at resolution (CR 611.2c): a token
// made afterwards is not indestructible.
//
// No simplification. Mass Casualty deck request (#3245).
func init() {
	Register(Spec{
		OracleID:     "8c1fda45-09f1-4b58-a487-e11c868e1357",
		Name:         "Ainok Strike Leader",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, thisOrYourCommanderAttacked,
				"Ainok Strike Leader — a tapped and attacking Goblin for each opponent", ainokGoblins)),
		},
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice this creature: Creature tokens you control gain indestructible until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    game.AbilityCost{SacrificeSelf: true},
			Effect: Do(GrantKeywordUntilEOT{
				Match:    And(Creature(), IsTokenPredicate(), YouControl()),
				Keywords: []string{"indestructible"},
				Label:    "Ainok Strike Leader — indestructible",
			}),
		}},
	})
}

// thisOrYourCommanderAttacked is "whenever you attack with this
// creature and/or your commander": the attacker is the source, or a
// commander the source's controller owns and attacks with.
func thisOrYourCommanderAttacked(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !attackDeclaredByYou(ev, source.Controller) {
		return false
	}
	if ev.CardID == source.InstanceID {
		return true
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCommander && c.Owner == source.Controller
}

func ainokGoblins(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, opp := range ctx.Opponents() {
		tmpl := RedGoblinToken()
		tmpl.Tapped = true
		tmpl.AttackingTarget = opp
		if err := (CreateToken{Controller: item.Controller, Template: tmpl, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
