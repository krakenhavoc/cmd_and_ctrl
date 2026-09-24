package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Agadeem's Awakening // Agadeem, the Undercrypt — modal double-faced
// card. This file is the FRONT face, Sorcery {X}{B}{B}{B}:
//
//	"Return from your graveyard to the battlefield any number of
//	 target creature cards that each have a different mana value X
//	 or less."
//
// The back face, Agadeem, the Undercrypt, is registered with the MDFC
// land cycle in mdfc_lands.go under "<oracle>#1".
//
// The card #1559 was opened for. Its clause has two halves and
// neither is a per-candidate predicate:
//
//   - "each have a different mana value" is a rule about the SET of
//     picks — EachDifferentManaValue. The announce gate refuses two
//     cards of one mana value (CR 601.2c), the picker greys the
//     second, the bot is never offered the pair.
//   - "X or less" reads the X announced at CR 601.2b, which a CardOK
//     predicate is never handed — WithManaValueAtMostX. The engine
//     binds it to the announced X at cast and again from the stack
//     item's X at resolution, and the picker narrows the legal set by
//     the X it collected.
//
// Resolution (CR 608.2b): a card exiled from the graveyard in
// response is skipped and the rest still return. Mana values of
// cards in a graveyard do not change, so the set rule cannot newly
// fail at resolution — but it is judged there all the same.
//
// The cards return under their owner's control — the caster's, since
// "from your graveyard" — in announce order, each through the ordinary
// reanimation path, as Eerie Ultimatum's do.
//
// "Any number" includes none: X=0 with no targets is a legal cast that
// does nothing, and a creature card of mana value 0 is a legal target
// at X=0.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "562d71b9-1646-474e-9293-55da6947a758",
		Name:         "Agadeem's Awakening",
		Completeness: CompletenessFull,
		Targets: TargetCardInGraveyard("any number of target creature cards that each have a different mana value X or less",
			Creature(), YouOwn()).WithCount(0, 0).EachDifferent(EachDifferentManaValue()).WithManaValueAtMostX(),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return returnLegalGraveyardTargetsToBattlefield(ctx)
		},
	})
}
