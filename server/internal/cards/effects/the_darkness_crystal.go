package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Darkness Crystal — Legendary Artifact {2}{B}{B}:
//
//	"Black spells you cast cost {1} less to cast.
//	 If a nontoken creature an opponent controls would die, instead
//	 exile it and you gain 2 life.
//	 {4}{B}{B}, {T}: Put target creature card exiled with The Darkness
//	 Crystal onto the battlefield tapped under your control with two
//	 additional +1/+1 counters on it."
//
// Three lines, three existing shapes (#3213, the Kuja deck):
//
//   - The discount is the Water Crystal's, in black: YourSpell and
//     ColoredSpell("B"). It only ever takes generic mana off.
//   - The replacement is Liesa's "if a creature an opponent controls
//     would die, exile it instead" narrowed to nontoken creatures,
//     with the exile linked to the Crystal (ReplacementEvent.ExiledWith,
//     CR 607.2a) and the 2 life gained as the replacement applies. The
//     creature never died (CR 700.4), so its dies-triggers do not fire.
//     A commander is exiled too and its owner then decides about the
//     command zone (CR 903.9a).
//   - The ability returns a creature card whose link names this Crystal
//     (the object, so a second Crystal's exiles are not offered). The
//     counters are an "enters with" on the entry event, so a counter
//     replacement sees them and an enters trigger finds them; the new
//     object is tapped and under YOUR control.
//
// The creature card is chosen as the ability resolves rather than when
// it is activated: the engine has no target clause for a card in exile
// yet (the picker only opens on the battlefield, the stack and
// graveyards), so the printed "target" is a choice. Nothing at the
// table can act on a card in exile in the meantime, which keeps this
// close, but it is declared. A card that left exile in response cannot
// be chosen.
func init() {
	Register(Spec{
		OracleID:     "660f370a-b25b-4ea5-b765-7abf96899c1c",
		Name:         "The Darkness Crystal",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The creature card to bring back is chosen when the ability resolves rather than when it's activated, so the ability can't be responded to with that choice in view.",
		},
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Black spells you cast cost {1} less to cast.",
				YourSpell(), ColoredSpell("B")),
		},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if !rfOpponentCreatureWouldDie(ev, g, src) {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && !c.IsToken()
			},
			Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
				ev.NewZone = game.ZoneExile
				ev.NewZoneOwner = uuid.Nil
				ev.ExiledWith = game.PermissionCardRef{ID: src.InstanceID, Epoch: src.ObjectEpoch}
				return g.ChangePlayerLifeForEffect(src.InstanceID, src.Controller, 2)
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "The Darkness Crystal: exile instead of dying, and gain 2 life",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{4}{B}{B}, {T}: Put target creature card exiled with The Darkness Crystal onto the battlefield tapped under your control with two additional +1/+1 counters on it.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{4}{B}{B}"), TapCost()),
			Effect:  darknessCrystalReturn,
		}},
	})
}

// darknessCrystalCandidates are the creature cards in exile whose
// "exiled with" link names this Crystal as the object it is (CR 400.7).
func darknessCrystalCandidates(g *game.Game, source uuid.UUID) []uuid.UUID {
	epoch := 0
	if c, ok := g.LookupCardForEffect(source); ok {
		epoch = c.ObjectEpoch
	}
	var ids []uuid.UUID
	if g.Exile == nil {
		return nil
	}
	for _, c := range g.Exile.Cards {
		if c.ExiledWith.ID != source || !c.IsCreature() {
			continue
		}
		if epoch != 0 && c.ExiledWith.Epoch != epoch {
			continue
		}
		ids = append(ids, c.InstanceID)
	}
	return ids
}

func darknessCrystalReturn(g *game.Game, item *game.StackItem) error {
	you := item.Controller
	ids := darknessCrystalCandidates(g, item.SourceCardID)
	if len(ids) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  you,
		Source:   item.SourceCardID,
		Question: "The Darkness Crystal — choose a creature card exiled with it to put onto the battlefield tapped under your control",
		Cards:    ids,
		Min:      1,
		Max:      1,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			for _, id := range picked {
				if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
					continue
				}
				if err := (ReturnFromExile{
					Target:     id,
					Controller: you,
					Tapped:     true,
					Counters:   map[string]int{game.CounterPlusOne: 2},
				}).Apply(NewContext(g, item)); err != nil {
					return err
				}
			}
			return nil
		},
	})
	return nil
}
