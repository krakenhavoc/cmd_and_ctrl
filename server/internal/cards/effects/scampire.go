package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scampire — Creature — Vampire Employee {2}{B}, 1/3:
//
//	"When Scampire enters, you get {TK}, then you may put a sticker on a
//	 creature card in your graveyard.
//	 {3}{B}: Return target creature card with a sticker on it from your
//	 graveyard to the battlefield. That creature gains haste. Exile it at
//	 the beginning of the next end step. Activate only as a sorcery."
//
// A sticker stays on a card that moves between public zones and comes off
// one that goes to a hidden zone (CR 123.5), so a graveyard holds stickered
// cards. A name or ability sticker applies there too (CR 123.6, 123.7),
// and the returned creature keeps the sticker onto the battlefield. The
// target is "stickered" (CR 123.4): any sticker on it now.
//
// The return is Whip of Erebos's: haste until end of turn, and the
// creature is exiled at the beginning of the next end step.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "64dbed80-b3a2-4a0f-afe4-c3d968173e50",
		Name:         "Scampire",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Scampire — you get {TK}, then you may put a sticker on a creature card in your graveyard",
				Do(GetTickets{N: 1}, PutASticker{On: StickerOnCreatureCardInYourGraveyard})),
		},
		Activated: []ActivatedAbility{{
			Label: "{3}{B}: Return target creature card with a sticker on it from your graveyard to the battlefield. That creature gains haste. Exile it at the beginning of the next end step. Activate only as a sorcery.",
			Cost:  ManaCost("{3}{B}"),
			Targets: TargetCardInGraveyard("target creature card with a sticker on it from your graveyard",
				Creature(), YouOwn(), Stickered()),
			SorcerySpeed: true,
			Effect:       scampireReturn,
		}},
	})
}

func scampireReturn(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	ctx := NewContext(g, item)
	id := item.Targets[0].ID
	if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield}).Apply(ctx); err != nil {
		return err
	}
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	if err := (GrantKeywordUntilEOT{
		Target:   id,
		Keywords: []string{"haste"},
		Label:    "Scampire — the returned creature has haste",
	}).Apply(ctx); err != nil {
		return err
	}
	return ScheduleDelayedTrigger{
		Label: "Scampire — exile the returned creature",
		Cards: []uuid.UUID{id},
		Body:  exileListedCardsBody,
	}.Apply(ctx)
}
