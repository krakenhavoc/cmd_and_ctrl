package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Motivated Pony — Creature — Horse {4}{G}, 3/3 (Tales of Middle-earth
// Commander, #5948):
//
//	"Trample, haste
//	 Whenever this creature attacks, attacking creatures get +1/+1 until end
//	 of turn. If a Food entered the battlefield under your control this turn,
//	 untap those creatures and they get an additional +2/+2 until end of turn."
//
// The creatures that are attacking as the trigger resolves (any player's, as
// printed) are the set (CR 611.2c): each gets +1/+1, and "those creatures" are
// that same set for the rest of the sentence. The Food clause is checked as
// the trigger resolves (CR 608.2), from the turn's tally of permanents that
// entered under your control (Food is an artifact subtype, so a Food token
// counts and so does a Food card); then each of those creatures untaps (an
// untapped creature stays untapped, and an attacker that untaps remains an
// attacker, CR 506.4) and gets +2/+2.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "44906d21-83cc-49ec-8505-4f7b020ac05f",
		Name:            "Motivated Pony",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Motivated Pony — attacking creatures get +1/+1 until end of turn; if a Food entered under your control this turn, untap them and they get +2/+2",
				motivatedPonyRally),
		},
	})
}

func motivatedPonyRally(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	attackers := AttackingCreatures(g)
	for _, id := range attackers {
		if err := (BoostUntilEOT{Target: id, Power: 1, Toughness: 1, Label: "Motivated Pony — +1/+1"}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	if g.EnteredWithSubtypeThisTurn(item.Controller, "Food") == 0 {
		return nil
	}
	if err := untapEach(ctx, attackers); err != nil {
		return err
	}
	for _, id := range attackers {
		if err := (BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: "Motivated Pony — an additional +2/+2"}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}
