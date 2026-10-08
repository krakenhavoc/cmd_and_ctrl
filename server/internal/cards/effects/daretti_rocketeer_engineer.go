package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Daretti, Rocketeer Engineer — Legendary Creature — Goblin Artificer
// {4}{R}, */5:
//
//	"Daretti's power is equal to the greatest mana value among
//	 artifacts you control.
//	 Whenever Daretti enters or attacks, choose target artifact card in
//	 your graveyard. You may sacrifice an artifact. If you do, return
//	 the chosen card to the battlefield."
//
// The power is a layer 7a characteristic-defining ability (Karn, Legacy
// Reforged's shape, power only). The trigger targets the graveyard card
// when it goes on the stack and re-checks it on resolution (CR 608.2b).
// "You may sacrifice an artifact" is a yes/no question and then the
// sacrifice prompt, and the return is the sacrifice's continuation, so
// it only happens once the artifact has really gone ("if you do").
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0eb1d539-6301-46b0-903e-487eda34253f",
		Name:         "Daretti, Rocketeer Engineer",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7A_CDA,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := 0
				for _, p := range g.BattlefieldCardsForEffect() {
					if p.Controller == source.Controller && p.IsArtifact() && p.ManaValue() > n {
						n = p.ManaValue()
					}
				}
				c.Power = n
			},
		}},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB, game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Targets: TargetCardInGraveyard("target artifact card in your graveyard", Artifact(), YouOwn()),
			Key:     "Daretti — sacrifice an artifact to return the chosen artifact card",
			Effect:  darettiSacrificeToReturn,
		}},
	})
}

// darettiSacrificeToReturn is the trigger body. Caller holds g.mu.
func darettiSacrificeToReturn(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	ctx := NewContext(g, item)
	if len(ctx.LegalTargets()) == 0 {
		return nil
	}
	hasArtifact := false
	for _, p := range g.BattlefieldCardsForEffect() {
		if p.Controller == item.Controller && p.IsArtifact() {
			hasArtifact = true
			break
		}
	}
	if !hasArtifact {
		return nil
	}
	return MayChoice{
		Question: "Daretti — sacrifice an artifact to return the chosen card?",
		YesLabel: "Sacrifice an artifact",
		NoLabel:  "Don't",
		OnYes:    darettiSacrificeThenReturn,
	}.Apply(ctx)
}

func darettiSacrificeThenReturn(ctx *Context) error {
	item := ctx.Item
	return ctx.Game.PlayerSacrificesThenForEffect(
		item.SourceCardID, item.Controller,
		sacrificeSpec("an artifact", Artifact()),
		"Daretti — sacrifice an artifact",
		1,
		func(g *game.Game, sacrificed game.PromptedSacrifices) error {
			if sacrificed.Count() == 0 {
				return nil
			}
			c := NewContext(g, item)
			for _, t := range c.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if z := g.FindCardZoneForEffect(t.ID); z == nil || z.Kind != game.ZoneGraveyard {
					return nil
				}
				return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield}.Apply(c)
			}
			return nil
		})
}
