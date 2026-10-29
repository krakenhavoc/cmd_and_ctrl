package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Daredevil, Man Without Fear — Legendary Creature — Human Hero
// {2}{R}{W}, 3/4 (Marvel Super Heroes, #4455):
//
//	"Vigilance, haste
//	 Radar Sense — You may look at the top card of your library any time.
//	 Whenever you attack, you may exile the top card of your library. If
//	 that card is a Hero card, Daredevil gets +2/+1 until end of turn. You
//	 may play that card this turn."
//
// Radar Sense is the private top-card designation (Realmwalker's
// LibraryTopOwner). "Whenever you attack" is ONE trigger per attack
// declaration (Adeline's OncePerBatch). The "you may" is the CR 603.5
// prompt asked as the trigger resolves; a yes exiles the top card with the
// this-turn play permission (a land may be played as the land drop), and
// the continuation, run once the card has landed in exile, boosts Daredevil
// if that card is a Hero card. Daredevil gets nothing if he has left.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:          "88aadf60-91f9-4932-a2bc-63365a139d36",
		Name:              "Daredevil, Man Without Fear",
		Completeness:      CompletenessFull,
		PrintedKeywords:   []string{"vigilance", "haste"},
		LibraryTopVisible: game.LibraryTopOwner,
		Triggered: []game.TriggeredAbility{
			Optional(OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller)
			}, "Daredevil — exile the top card of your library; you may play it this turn", func(g *game.Game, item *game.StackItem) error {
				return b12ImpulseExileForTurnThen(g, item, 1, daredevilExiled)
			})), "Daredevil — exile the top card of your library?"),
		},
	})
}

// daredevilExiled is "if that card is a Hero card, Daredevil gets +2/+1
// until end of turn", run once the exiled card has landed. An on-disk key.
var daredevilExiled = CardTail("daredevil-man-without-fear/hero-boost", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 || !onBattlefield(ctx.Game, ctx.Source()) {
		return nil
	}
	card, ok := ctx.Game.LookupCardForEffect(o.IDs[0])
	if !ok || !card.HasSubtype(ctx.TextSubtype("Hero")) {
		return nil
	}
	return BoostUntilEOT{Target: ctx.Source(), Power: 2, Toughness: 1, Label: "Daredevil — +2/+1"}.Apply(ctx)
})
