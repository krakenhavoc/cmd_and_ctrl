package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Party Dude — {G} Enchantment — Class (Teenage Mutant Ninja Turtles,
// #4457):
//
//	"(Gain the next level as a sorcery to add its ability.)
//	 When this Class enters, each player creates a Food token.
//	 {1}{G}: Level 2
//	 Whenever an artifact an opponent controls is put into a graveyard
//	 from the battlefield, draw a card.
//	 {4}{G}: Level 3
//	 Whenever one or more of your opponents are attacked, up to one target
//	 attacking creature gets +X/+X until end of turn, where X is the
//	 number of cards in your hand."
//
// ADR 0071 Class machinery (Innkeeper's Talent): LevelUp for each level,
// and each level's trigger gated with AtLevel, so it exists only from
// that level on (CR 716.2a).
//
//   - Level 1: each player, in turn order from the active player, creates
//     a Food (Descent into Avernus's loop).
//   - Level 2: Disciple of the Vault's "an artifact is put into a
//     graveyard from the battlefield", narrowed to one an opponent
//     controlled as it last existed (leftUnderControlOf, CR 603.10a). A
//     token counts: it goes to the graveyard before it ceases to exist.
//   - Level 3: "one or more of your opponents are attacked" is one
//     trigger per attack declaration (OncePerBatch), whoever attacks,
//     when at least one creature attacks a player who is your opponent.
//     An attack on a planeswalker or a battle is not an attack on its
//     player (CR 508.3b), so it alone does not trigger. The target is any
//     attacking creature, chosen as the trigger goes on the stack; X is
//     the cards in your hand as it resolves (Bast, Panther Goddess's
//     shape).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8df7a58c-053f-4ead-a778-2747718e5f10",
		Name:         "Party Dude",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Party Dude — each player creates a Food token", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, p := range seatsFromActive(g) {
					if err := (CreateToken{Controller: p, Template: FoodToken(), N: 1}).Apply(ctx.asGroupMember()); err != nil {
						return err
					}
				}
				return nil
			}),
			AtLevel(2, On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !b23ArtifactPutIntoGraveyardFromBattlefield(ev, g) {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				if !ok {
					return false
				}
				controller := leftUnderControlOf(ev, c)
				return controller != uuid.Nil && controller != source.Controller
			}, "Party Dude — draw a card", Do(DrawCards{N: 1}))),
			AtLevel(3, OncePerBatch(Targeting(
				On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					p := attackedPlayer(g, ev)
					return p != uuid.Nil && p != source.Controller
				}, "Party Dude — up to one target attacking creature gets +X/+X, where X is the number of cards in your hand",
					partyDudePump),
				UpToOneTargetCreature("up to one target attacking creature", AttackingCreature())))),
		},
		Activated: []ActivatedAbility{
			LevelUp(2, ManaCost("{1}{G}")),
			LevelUp(3, ManaCost("{4}{G}")),
		},
	})
}

// partyDudePump is level 3's resolution: +X/+X until end of turn on the
// target, X the cards in your hand now.
func partyDudePump(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	x := 0
	if p := g.PlayerByIDForEffect(item.Controller); p != nil && p.Hand != nil {
		x = p.Hand.Size()
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return BoostUntilEOT{Target: t.ID, Power: x, Toughness: x, Label: "Party Dude — +X/+X"}.Apply(ctx)
	}
	return nil
}
