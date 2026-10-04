package effects

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// emblem.go is the catalog half of CR 114 (#623, ADR 0064).
//
// A card that prints "You get an emblem with [ability]" declares the
// emblem ONCE, on its own Spec, next to the ability that creates it:
//
//	Register(Spec{
//	    OracleID: "05e6b243-…",
//	    Name:     "Elspeth, Sun's Champion",
//	    Emblem: &EmblemSpec{
//	        Label: "Elspeth, Sun's Champion emblem",
//	        Text:  "Creatures you control get +2/+2 and have flying.",
//	        Static: []game.StaticAbility{ … },
//	    },
//	    Activated: []ActivatedAbility{{
//	        Label:  "−7: You get an emblem with …",
//	        Cost:   LoyaltyCost(-7),
//	        Effect: func(g *game.Game, item *game.StackItem) error {
//	            return CreateEmblem{}.Apply(NewContext(g, item))
//	        },
//	    }},
//	})
//
// Register builds a SECOND game.CardDef out of that slot and files it
// in the same `defs` map cards use, keyed game.EmblemKey(OracleID) —
// "emblem:<oracle id>". That is the whole trick: the emblem's
// abilities reach the layer pass and the trigger harvester through
// exactly the CatalogStaticAbilities / CatalogTriggers hooks a
// battlefield permanent's do, so an emblem's abilities are written in
// the same vocabulary and there is no emblem dialect.
//
// It goes into `defs` and NOT into `registry`, which is what All()
// returns. An emblem is not a card (CR 114.4) and must not be counted
// by the card-coverage census.

// EmblemSpec is the emblem a card's ability creates (CR 114.5).
//
// Label and Text are what a player reads: the chip on the board and
// its hover text. They are presentation, looked up fresh on every
// wire projection, so a wording fix in a card file reaches a game
// already in progress.
//
// Static and Triggered are the emblem's abilities, in the same shapes
// a permanent uses. An emblem's static applies from the command zone
// (CR 114.3) and its AppliesTo / Apply receive the emblem object as
// `source`, so "creatures you control" is the same `target.Controller
// == source.Controller` it is on an anthem. An emblem's trigger goes
// on the stack, takes its CR 603.3d target when it is put there, and
// can be responded to, like every other trigger.
type EmblemSpec struct {
	// Label names the emblem on the board and in the log. Convention
	// is "<source card> emblem" — the emblem has no name of its own
	// (CR 114.1), so this is a label, not a characteristic.
	Label string

	// Text is the emblem's printed ability, verbatim, with no
	// surrounding quotes. Shown on hover.
	Text string

	// Static and Triggered are the emblem's abilities.
	Static    []game.StaticAbility
	Triggered []game.TriggeredAbility

	// UntapStep and DrawStep are the emblem's contributions to the
	// two turn-based actions CR 114.3 also runs from the command zone
	// (#1315): "untap all permanents you control during each
	// opponent's untap step" and "you draw a card during each
	// opponent's draw step" (Teferi, Who Slows the Sunset). Neither
	// uses the stack, which is why they are not Triggered entries —
	// see game/untap.go and game/draw_step.go for the argument.
	//
	// At least one ability slot (these two, Static, Triggered,
	// ActivationTimings or one of the rule gates below) must be
	// non-empty — an emblem with no abilities has no characteristics
	// at all and would be an object nothing can observe.
	UntapStep []game.UntapStepPermission
	DrawStep  []game.DrawStepPermission

	// ActivationTimings are the emblem's per-player activation-TIMING
	// statements (#1275, CR 602.5d / CR 606.3 / CR 101.1) — the
	// `Spec.ActivationTimings` slot one zone over. Teferi, Temporal
	// Archmage's −10: "You may activate loyalty abilities of
	// planeswalkers you control on any player's turn any time you
	// could cast an instant."
	//
	// The same `game.ActivationTiming` a permanent declares, built by
	// the same constructors and checked by the same Register guard.
	// `Covers` receives the EMBLEM as `q.Source`, so "you" is
	// `q.Source.Controller` — the emblem's owner, which never changes
	// (CR 114.2) — exactly as it is the permanent's controller on
	// Leonin Shikari.
	ActivationTimings []game.ActivationTiming

	// CastRestrictions, LandPlayRestrictions, GameEndGates and UntapCaps
	// are the emblem's rule gates (ADR 0109 §5, #1899): the Spec slots of
	// the same names one zone over, built by the same constructors and
	// checked by the same Register guards. Each gate that reads the Spec
	// slot also walks every seat's emblems (CR 114.4), with the emblem
	// as the source, so "your opponents" is the emblem's owner's
	// opponents (CR 114.2):
	//
	//	CastRestrictions      Narset Transcendent's "Your opponents can't
	//	                      cast noncreature spells" (the cast gate)
	//	LandPlayRestrictions  the land-play gate; no printed emblem yet
	//	GameEndGates          Gideon of the Trials' "you can't lose the
	//	                      game and your opponents can't win the game"
	//	UntapCaps             Dovin Baan's "Your opponents can't untap
	//	                      more than two permanents during their untap
	//	                      steps"
	CastRestrictions     []game.CastRestriction
	LandPlayRestrictions []game.LandPlayRestriction
	GameEndGates         []game.GameEndGate
	UntapCaps            []game.UntapCap

	// BlockRules are the emblem's CR 509.1b block restrictions with a
	// parameter (ADR 0114 §5): the Spec slot of the same name one zone
	// over, built by the same constructors (block_rules.go). The
	// block-rule walk reads every seat's emblems after the battlefield,
	// with the emblem as the rule's source, so a scope's "your" is the
	// emblem's owner: the Ring's "your Ring-bearer can't be blocked by
	// creatures with greater power".
	BlockRules []game.BlockRule

	// Lines are the emblem's abilities line by line, each with the
	// count it is gained at, for an emblem whose abilities switch on one
	// by one (the Ring, CR 701.54c). The wire's text for it is derived
	// from the lines it has right now. Empty for every card's emblem.
	Lines []game.EmblemLine
}

// buildEmblemDef projects an EmblemSpec into the CardDef the engine
// reads for the emblem object. Deliberately narrow: an emblem has no
// cost, no targets of its own, no printed keywords and no resolve
// hook, so the slots it does not fill stay nil rather than being
// copied from the card that made it.
func buildEmblemDef(e EmblemSpec) *game.CardDef {
	return &game.CardDef{
		Static:            e.Static,
		Triggered:         e.Triggered,
		UntapStep:         e.UntapStep,
		DrawStep:          e.DrawStep,
		ActivationTimings: e.ActivationTimings,
		// ADR 0109 §5: the rule gates, which the gates read off this def
		// through the same catalog hooks a permanent's go through.
		CastRestrictions:     e.CastRestrictions,
		LandPlayRestrictions: e.LandPlayRestrictions,
		GameEndGates:         e.GameEndGates,
		UntapCaps:            e.UntapCaps,
		// ADR 0114 §5: read through CatalogBlockRules by the
		// block-rule walk's emblem half.
		BlockRules: e.BlockRules,
		Emblem:     &game.EmblemDef{Label: e.Label, Text: e.Text, Lines: e.Lines},
	}
}

// checkEmblemSpec is the registration guard. Every failure here is a
// card file that is wrong in a way no test of the card would catch,
// so it fails at boot.
func checkEmblemSpec(name string, e *EmblemSpec) {
	if e == nil {
		return
	}
	if e.Label == "" {
		panic("effects.Register: " + name + " declares an Emblem with no Label — the board has nothing to call it")
	}
	if e.Text == "" {
		panic("effects.Register: " + name + " declares an Emblem with no Text — a player reading the chip learns nothing")
	}
	if len(e.Static) == 0 && len(e.Triggered) == 0 && len(e.UntapStep) == 0 && len(e.DrawStep) == 0 &&
		len(e.ActivationTimings) == 0 && len(e.CastRestrictions) == 0 && len(e.LandPlayRestrictions) == 0 &&
		len(e.GameEndGates) == 0 && len(e.UntapCaps) == 0 && len(e.BlockRules) == 0 {
		panic("effects.Register: " + name + " declares an Emblem with no abilities — CR 114.1 says an emblem has nothing else")
	}
	checkActivationTimings(name+" emblem", e.ActivationTimings)
	checkCastRestrictions(name+" emblem", e.CastRestrictions)
	checkLandPlayRestrictions(name+" emblem", e.LandPlayRestrictions)
	// ADR 0107 §3: the stack step gathers statics over spells from the
	// battlefield only, so one on an emblem would be dead text.
	for i, s := range e.Static {
		if s.AffectsSpells {
			panic(fmt.Sprintf("effects.Register: %s emblem static %d affects spells — only a battlefield permanent's static over spells is applied", name, i))
		}
	}
}

// CreateEmblem is CR 114.5's "you get an emblem". The emblem is the
// one the SOURCE card declares, so there is nothing to pass: the zero
// value creates the resolving card's emblem for the item's
// controller, which is what every printed emblem clause says.
//
// A GRANTED ability's emblem is the grantor's (ADR 0109 §2, owner
// decision 2). Teferi's Talent gives the planeswalker it enchants
// "[−12]: You get an emblem with '…'"; the walker is the ability's
// source and pays the 12, but the emblem's text is printed on the
// Talent, so the zero value files it under the Talent's key
// (Context.Grantor) and the Talent's Spec declares the emblem (CR
// 114.2).
//
// Player and Source override those defaults for the case that does
// not exist yet — an effect that hands somebody else's emblem to
// somebody else — and are here so the primitive does not have to grow
// a second constructor the day one is printed.
type CreateEmblem struct {
	Player uuid.UUID
	Source uuid.UUID
}

func (c CreateEmblem) Apply(ctx *Context) error {
	player := c.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	source := c.Source
	if source == uuid.Nil {
		source = ctx.Grantor()
	}
	if source == uuid.Nil {
		source = ctx.Source()
	}
	return ctx.Game.CreateEmblemForEffect(player, source)
}
