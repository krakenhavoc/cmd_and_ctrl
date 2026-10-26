package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ambassador Blorpityblorpboop — Legendary Creature — Alien Advisor Guest
// {3}{G}{U}, 3/3:
//
//	"When Ambassador Blorpityblorpboop enters, you get {TK}{TK}{TK}, then
//	 you may put a sticker on a nonland permanent you own.
//	 At the beginning of each combat, you may have Ambassador
//	 Blorpityblorpboop's base power become equal to the total power of all
//	 stickers on permanents you control and its base toughness become equal
//	 to those stickers' total toughness."
//
// "The power of a sticker" is the printed number on a power and toughness
// sticker; no other sticker has one (CR 123.8a). The totals are read as the
// ability resolves and locked in (CR 611.2c), and the base power and
// toughness are set for as long as it stays on the battlefield (CR 611.2a:
// no duration is stated), a layer 7b effect (CR 613.4b) that a later
// effect or sticker with a later timestamp overrides and that +1/+1
// counters and pumps still modify. It is asked each combat, so a player
// can reset it, or keep what it has.
//
// No simplification.
func init() {
	reset := Optional(
		AtBeginningOfEachCombat("Ambassador Blorpityblorpboop — you may have its base power and toughness become the total power and toughness of all stickers on permanents you control",
			ambassadorSetBase),
		"Have Ambassador Blorpityblorpboop's base power and toughness become the total power and toughness of the stickers on permanents you control?")
	Register(Spec{
		OracleID:     "183a9488-e21f-448b-8b69-e53b56fa0a56",
		Name:         "Ambassador Blorpityblorpboop",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Ambassador Blorpityblorpboop — you get {TK}{TK}{TK}, then you may put a sticker on a nonland permanent you own",
				Do(GetTickets{N: 3}, PutASticker{})),
			reset,
		},
	})
}

// stickerPTTotals is the total power and the total toughness printed on
// the power and toughness stickers on the permanents `you` control.
func stickerPTTotals(g *game.Game, you uuid.UUID) (power, toughness int) {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != you {
			continue
		}
		for _, s := range c.Stickers {
			if s.Print.Kind == game.StickerPT {
				power += s.Print.Power
				toughness += s.Print.Toughness
			}
		}
	}
	return power, toughness
}

func ambassadorSetBase(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if g.AbilitySourceGoneForEffect(item) {
		return nil
	}
	p, t := stickerPTTotals(g, item.Controller)
	return ScopedEffectFor{
		Target:   item.SourceCardID,
		Mods:     []game.Mod{game.SetBasePowerMod(p), game.SetBaseToughnessMod(t)},
		Duration: g.PinnedTo(game.IndefiniteDuration(), item.SourceCardID),
		Label:    "Ambassador Blorpityblorpboop — base power and toughness become the stickers' totals",
	}.Apply(ctx)
}
