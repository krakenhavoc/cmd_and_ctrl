package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Dawning Archaic — Legendary Creature — Avatar {10}, 7/7 (Secrets of
// Strixhaven, #4456):
//
//	"This spell costs {1} less to cast for each instant and sorcery card in
//	 your graveyard.
//	 Reach
//	 Whenever The Dawning Archaic attacks, you may cast target instant or
//	 sorcery card from your graveyard without paying its mana cost. If that
//	 spell would be put into your graveyard, exile it instead."
//
// The reduction is a self cost modifier (Cyan, Vengeful Samurai's shape) that
// counts the caster's instant and sorcery cards in their graveyard when the
// spell is cast (CR 601.2f); it only lowers the generic part. The attack
// trigger is Torrential Gearhulk's free cast: a {0}, any-timing, cast-only
// permission over the target card, offered during the resolution (CR 608.2g),
// with the permission's ExileInsteadOfGraveyard making the spell go to exile
// whether it resolves, is countered or fizzles. The target is chosen as the
// trigger goes on the stack and re-checked as it resolves (CR 608.2b);
// declining the cast leaves the card where it is.
//
// No simplification.
func init() {
	attack := WheneverThisAttacks("The Dawning Archaic — you may cast target instant or sorcery card from your graveyard without paying its mana cost",
		freeCastTargetFromYourGraveyardExilingIt("The Dawning Archaic"))
	attack.Targets = TargetCardInGraveyard("target instant or sorcery card from your graveyard",
		Or(Instant(), Sorcery()), YouOwn())
	Register(Spec{
		OracleID:        "a3e10b9b-9349-4b44-a46c-c825293dbd05",
		Name:            "The Dawning Archaic",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(socB37InstantAndSorceryCardsInYourGraveyard,
				"This spell costs {1} less to cast for each instant and sorcery card in your graveyard."),
		},
		Triggered: []game.TriggeredAbility{attack},
	})
}
