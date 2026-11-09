package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dion, Bahamut's Dominant // Bahamut, Warden of Light (Final Fantasy,
// #4462). Front, Legendary Creature — Human Noble Knight {3}{W}, 3/3:
//
//	"Dragonfire Dive — During your turn, Dion and other Knights you
//	 control have flying.
//	 When Dion enters, create a 2/2 white Knight creature token.
//	 {4}{W}{W}, {T}: Exile Dion, then return it to the battlefield
//	 transformed under its owner's control. Activate only as a sorcery."
//
// Back, Legendary Enchantment Creature — Saga Dragon, 5/5:
//
//	"(As this Saga enters and after your draw step, add a lore counter.)
//	 I, II — Wings of Light — Put a +1/+1 counter on each other creature
//	 you control. Those creatures gain flying until end of turn.
//	 III — Gigaflare — Destroy target permanent. Exile Bahamut, then
//	 return it to the battlefield (front face up).
//	 Flying"
//
// Crystal Fragments' flip (ExileAndReturnTransformed): two zone changes,
// a new object (CR 400.7) that enters as a Saga with its first lore
// counter. Dragonfire Dive grants flying to Dion himself whether or not
// he is a Knight (the ruling) and to the Knights the controller controls,
// on that player's own turn only.
//
// Wings of Light picks the creatures once, as it resolves (CR 611.2c): the
// ones that got a counter are the ones that fly. Gigaflare is the chapter
// III shape of Fable of the Mirror-Breaker with a target: an illegal
// target stops the whole ability, so Bahamut is neither exiled nor
// returned and the Saga is sacrificed by state-based action (the
// ruling). Exiling and returning it flips it, from its back face to its
// front one, and Dion enters again, making another Knight.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     dionOracleID,
		Name:         "Dion, Bahamut's Dominant",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{dionDragonfireDive()},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Dion, Bahamut's Dominant — create a 2/2 white Knight creature token",
				Do(CreateToken{Template: TokenCard("2/2 white Knight"), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:        "{4}{W}{W}, {T}: Exile Dion, then return it to the battlefield transformed under its owner's control. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{4}{W}{W}"), TapCost()),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return ExileAndReturnTransformed{Target: item.SourceCardID}.Apply(NewContext(g, item))
			},
		}},
	})
	Register(Spec{
		OracleID:        dionOracleID + "#1",
		Name:            "Bahamut, Warden of Light",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel("Bahamut, Warden of Light", 1, "Wings of Light — +1/+1 counter on each other creature you control; they gain flying until end of turn"), dionWingsOfLight),
			ChapterTrigger(2, SagaChapterLabel("Bahamut, Warden of Light", 2, "Wings of Light — +1/+1 counter on each other creature you control; they gain flying until end of turn"), dionWingsOfLight),
			ChapterTriggerTargeting(3, SagaChapterLabel("Bahamut, Warden of Light", 3, "Gigaflare — destroy target permanent; exile Bahamut, then return it front face up"),
				TargetPermanent("target permanent"), dionGigaflare),
		},
	})
}

// dionOracleID is shared by both faces.
const dionOracleID = "4a0cc2a2-3711-473a-b2c5-b1e52b34e4fb"

// dionDragonfireDive is "During your turn, Dion and other Knights you
// control have flying": a layer-6 grant over the source and the Knights
// its controller controls, while that player is the active player.
func dionDragonfireDive() game.StaticAbility {
	ab := KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
		if target.Controller != source.Controller || !target.IsCreature() || !isActivePlayer(g, source.Controller) {
			return false
		}
		return target.InstanceID == source.InstanceID || target.HasSubtype(source.TextSubtype("Knight"))
	}, "flying")
	ab.Label = "Dragonfire Dive — During your turn, Dion and other Knights you control have flying."
	return ab
}

// dionWingsOfLight is chapters I and II: a counter on, and flying for,
// each other creature its controller controls.
func dionWingsOfLight(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []game.Card
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID != item.SourceCardID && c.Controller == item.Controller && c.IsCreature() {
			ids = append(ids, c)
		}
	}
	for _, c := range ids {
		if err := (AddCounter{Target: c.InstanceID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
		if err := (GrantKeywordUntilEOT{Target: c.InstanceID, Keywords: []string{"flying"}, Label: "Bahamut, Warden of Light — Wings of Light: flying until end of turn"}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}

// dionGigaflare is chapter III: destroy the target, then exile Bahamut
// and return it front face up.
func dionGigaflare(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (DestroyTarget{Target: id}).Apply(ctx); err != nil {
		return err
	}
	return ChapterExileAndReturnTransformed(g, item)
}
