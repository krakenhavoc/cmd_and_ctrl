package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Seifer Almasy — Legendary Creature — Human Knight {3}{R}, 3/4 (Final
// Fantasy, #4462):
//
//	"Whenever a creature you control attacks alone, it gains double
//	 strike until end of turn.
//	 Fire Cross — Whenever Seifer Almasy deals combat damage to a
//	 player, you may cast target instant or sorcery card with mana value
//	 3 or less from your graveyard without paying its mana cost. If that
//	 spell would be put into your graveyard, exile it instead."
//
// The first ability is Black Widow, Double Agent's lone-attacker read with
// double strike (granted at declaration, before blockers). Fire Cross is
// Daring Waverider's free cast from the graveyard, exiling the spell
// whether it resolves, is countered or fizzles, offered as the trigger
// resolves (CR 608.2g), over a smaller mana value.
//
// No simplification.
func init() {
	fireCross := Targeting(
		WheneverThisDealsCombatDamageToAPlayer("Seifer Almasy — Fire Cross: you may cast target instant or sorcery card with mana value 3 or less from your graveyard without paying its mana cost",
			freeCastTargetFromYourGraveyardExilingIt("Seifer Almasy")),
		TargetCardInGraveyard("target instant or sorcery card with mana value 3 or less from your graveyard",
			Or(Instant(), Sorcery()), ManaValueLE(3), YouOwn()))
	Register(Spec{
		OracleID:     "7088a901-0489-41c3-9f2f-633939514de8",
		Name:         "Seifer Almasy",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, mshB1YouAttackAlone,
				"Seifer Almasy — the lone attacker gains double strike until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					attacker := item.Trigger.Event.CardID
					if !onBattlefield(g, attacker) {
						return nil
					}
					return GrantKeywordUntilEOT{
						Target:   attacker,
						Keywords: []string{"double strike"},
						Label:    "Seifer Almasy",
					}.Apply(NewContext(g, item))
				}),
			fireCross,
		},
	})
}
