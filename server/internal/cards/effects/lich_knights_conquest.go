package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lich-Knights' Conquest — Sorcery {4}{B} (EDHREC rank 2596):
//
//	"Sacrifice any number of artifacts, enchantments, and/or tokens.
//	 Return that many creature cards from your graveyard to the
//	 battlefield."
//
// Treasures and Clues into a mass reanimation. The two choices the
// card asks for — which permanents to sacrifice, which creature
// cards to return — are made in the opposite order from the printed
// one, because that is the pair of prompts the engine has: the
// creature cards are the spell's TARGET clause ("any number of
// target creature cards in your graveyard", picked at cast), and the
// sacrifices are that many sacrifice prompts at resolution, each the
// caster's own pick among their artifacts, enchantments and tokens
// (the b17PlayerSacrificesN shape, one permanent per prompt). The
// prompts are answered BEFORE the creatures return, so their option
// lists were computed without the returning cards — an artifact
// creature that just came back is never offered as the price of its
// own return. Fewer eligible permanents than chosen cards returns
// only as many as can be paid for; a card that left the graveyard
// in response is skipped (CR 608.2b).
//
// "THAT MANY" is how many were sacrificed (#1019, ADR 0013 §5x). It
// used to be how many prompts were QUEUED, read on the line after the
// last one went up: the creatures came back before anybody had picked
// a Treasure, and a prompt the engine withdrew — its token gone by
// the time the seat got to it — still bought a creature card.
//
// DECLARED SIMPLIFICATION, weaker than printed: the order of the
// choices. Choosing the creatures at cast means opponents see the
// picks and can respond to them, and — the printed card's best
// trick — a nontoken artifact or enchantment creature sacrificed to
// the spell cannot be one of the cards it returns, since it was not
// in the graveyard when the targets were chosen. Never stronger: the
// count is capped by what the caster can actually sacrifice.
func init() {
	Register(Spec{
		OracleID:     "b0e533dd-baf2-482a-b9e1-872eb5e47439",
		Name:         "Lich-Knights' Conquest",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"You pick the creature cards to return as you cast the spell and the permanents to sacrifice as it resolves — so a creature sacrificed to it can't be one of the cards returned, and opponents can respond to the picks.",
		},
		Targets: TargetCardInGraveyard("any number of target creature cards in your graveyard", YouOwn(), Creature()).WithCount(1, 0),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			chosen := 0
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					chosen++
				}
			}
			eligible := len(MatchingBattlefield(ctx, And(YouControl(), Or(Artifact(), Enchantment(), IsTokenPredicate()))))
			n := chosen
			if eligible < n {
				n = eligible
			}
			if n <= 0 {
				return nil
			}
			// "That many" is how many were SACRIFICED, not how many
			// prompts went up (#1019, ADR 0013 §5x). The run's
			// continuation waits for every one of the n prompts and
			// for the permanents they name to finish moving, so a
			// prompt withdrawn because its token had already gone, and
			// a sacrifice the CR 614 window cancelled, no longer buy a
			// creature card back.
			controller := ctx.Controller()
			return ctx.Game.PlayerSacrificesThenForEffect(
				ctx.Source(), controller,
				sacrificeSpec("an artifact, enchantment, or token", Or(Artifact(), Enchantment(), IsTokenPredicate())),
				"Lich-Knights' Conquest — sacrifice an artifact, enchantment, or token",
				n,
				func(g *game.Game, sacrificed game.PromptedSacrifices) error {
					return b24ReturnGraveyardTargetsToBattlefield(NewContext(g, item), sacrificed.Count())
				})
		},
	})
}
