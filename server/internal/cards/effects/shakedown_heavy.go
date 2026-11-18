package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shakedown Heavy — Creature — Ogre Warrior {2}{B}, 6/4 (Streets of New
// Capenna, #5491):
//
//	"Menace
//	 Whenever this creature attacks, defending player may have you draw a
//	 card. If they do, untap this creature and remove it from combat."
//
// The defending player is the player behind the attack (the controller of
// a planeswalker or protector of a battle it attacks, CR 506.3d). They are
// asked once as the trigger resolves (Combustible Gearhulk's shape), and
// on a yes the Heavy's controller draws and the Heavy is untapped and
// removed from combat (CR 506.4). A player a draw limit stops from drawing
// is not asked, which is a no. If the Heavy has left the battlefield, the
// draw still happens and nothing else does; a Heavy that has left combat
// is untapped all the same.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "77b24f10-93af-44b9-af78-3a2f5ab32251",
		Name:            "Shakedown Heavy",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Shakedown Heavy — defending player may have you draw a card; if they do, untap this creature and remove it from combat",
				shakedownHeavyAsk),
		},
	})
}

func shakedownHeavyAsk(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	defender := b17DefendingPlayer(g, item.Trigger.Event)
	if g.PlayerByIDForEffect(defender) == nil {
		return nil
	}
	return MayChoice{
		Player:   defender,
		Question: "Shakedown Heavy — let its controller draw a card? (It is untapped and removed from combat.)",
		Draws:    1, Drawer: ctx.Controller(),
		YesLabel: "They draw a card; untap it and remove it from combat",
		NoLabel:  "It stays in combat",
		Then:     shakedownHeavyAnswer,
	}.Apply(ctx)
}

var shakedownHeavyAnswer = MayYes("may/shakedown-heavy-draw", func(ctx *Context, _ game.ConfirmAnswer) error {
	if err := (DrawCards{Player: ctx.Controller(), N: 1}).Apply(ctx); err != nil {
		return err
	}
	id := ctx.Item.SourceCardID
	if !onBattlefield(ctx.Game, id) || sourceIsNewObject(ctx.Game, ctx.Item) {
		return nil
	}
	if err := (UntapTarget{Target: id}).Apply(ctx); err != nil {
		return err
	}
	return ctx.Game.RemoveFromCombatForEffect(id)
})
