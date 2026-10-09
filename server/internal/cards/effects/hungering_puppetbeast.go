package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hungering Puppetbeast — Artifact Creature — Beast Construct
// {3}{G}{G}, 5/5:
//
//	"When this creature enters, create a Heartwood token. (It's a red
//	 and green artifact with "{T}: Add {R} or {G}.")
//	 {1}, Sacrifice another artifact: Put a +1/+1 counter on this
//	 creature. It gains your choice of trample, hexproof, or haste until
//	 end of turn."
//
// "Your choice of" is a modal activated ability (ADR 0065, Cankerbloom's
// shape): the keyword is picked as the ability is activated, and every
// mode also puts the counter. The {1} and the sacrifice are paid up
// front.
//
// No simplification.
func init() {
	grant := func(keyword string) func(item *game.StackItem, ctx *Context, occurrence int) error {
		return func(item *game.StackItem, ctx *Context, _ int) error {
			self := item.SourceCardID
			if !onBattlefield(ctx.Game, self) {
				return nil
			}
			if err := (AddCounter{Target: self, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{
				Target:   self,
				Keywords: []string{keyword},
				Label:    "Hungering Puppetbeast — " + keyword + " until end of turn",
			}.Apply(ctx)
		}
	}
	Register(Spec{
		OracleID:     "690ae865-87bd-46b2-8e65-c53bd80c1a18",
		Name:         "Hungering Puppetbeast",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Hungering Puppetbeast — create a Heartwood token",
				Do(CreateToken{Template: HeartwoodToken(), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label: "{1}, Sacrifice another artifact: Put a +1/+1 counter on this creature. It gains your choice of trample, hexproof, or haste until end of turn.",
			Cost:  Plus(ManaCost("{1}"), SacrificeAnotherN(1, "another artifact", Artifact())),
			Modes: ChooseOne(
				ModeDoing("Put a +1/+1 counter on it. It gains trample until end of turn.", nil, grant("trample")),
				ModeDoing("Put a +1/+1 counter on it. It gains hexproof until end of turn.", nil, grant("hexproof")),
				ModeDoing("Put a +1/+1 counter on it. It gains haste until end of turn.", nil, grant("haste")),
			),
		}},
	})
}
