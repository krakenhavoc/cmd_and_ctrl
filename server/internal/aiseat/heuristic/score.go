package heuristic

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// score.go is the board-state evaluation the whole policy rests on:
// one number per seat for "how well is this seat doing", and one
// number per table for "how well am I doing relative to everyone
// else". Everything it reads comes off protocol.GameView — the same
// filtered projection a human at this seat receives — so the
// evaluation can never see a card the seat is not allowed to see.
// An opponent's hand contributes its COUNT and nothing else, which
// is exactly the information a player at the table has.

// Weights are the evaluation's tunable constants. DefaultWeights is
// what the `heuristic` tier ships with; the struct is exported so a
// future tier (or a test) can perturb one term without forking the
// function. Units are arbitrary but consistent: one point is
// roughly "one point of creature power".
type Weights struct {
	// Life is per point of life. Commander starts at 40, so life is
	// deliberately cheap per point — 40 life is not worth more than
	// a board.
	Life float64
	// LifeDanger is the extra penalty per point of life below
	// DangerLife, applied quadratically. Being at 4 is much worse
	// than twice as bad as being at 8.
	LifeDanger float64
	// DangerLife is where LifeDanger starts to bite.
	DangerLife int
	// Hand is per card in hand — cards are resources.
	Hand float64
	// Library is per card left in the library: decking is a real
	// loss condition, but a distant one.
	Library float64

	// Power and Toughness value a creature's printed-plus-layered
	// stats. Power is worth more than toughness.
	Power     float64
	Toughness float64
	// Keyword scales the keyword table in keywordBonus.
	Keyword float64
	// TappedCreature and SickCreature are multipliers on a creature
	// that cannot block or cannot act yet.
	TappedCreature float64
	SickCreature   float64
	// FrozenCreature further discounts a tapped creature that will miss
	// its controller's next untap step.
	FrozenCreature float64
	// CantAttack, CantBlock and CantActivate are multipliers on a
	// creature carrying the S24 restriction of that name (ADR 0045,
	// `CardView.Restrictions`). They compose, so Pacifism applies the
	// first two and Arrest all three, and they are what make a
	// removal Aura worth casting: without them a pacified creature
	// keeps its full value and the bot sees no gain (#727).
	//
	// "cant_be_blocked" is deliberately absent from the table. It is
	// carried on the attacker but restricts the DEFENDER's options —
	// a Whispersilk Cloak makes its host better, not worse — so
	// discounting the host for it would have the sign backwards.
	CantAttack   float64
	CantBlock    float64
	CantActivate float64

	// Permanent is the flat value of a non-creature, non-land
	// permanent. Planeswalker adds on top of it, and Loyalty values
	// each loyalty counter.
	Permanent    float64
	Planeswalker float64
	Loyalty      float64

	// AttachedAura and AttachedEquipment are what an attached BUFF is
	// worth on its OWN line, once its host already carries the boost
	// (#727 — see AttachmentRole). An Aura is worth nearly nothing
	// there: the value is the +2/+2 on the creature, and the Aura
	// follows it to the graveyard. An Equipment keeps a real residual
	// because it survives its host and can be moved to the next
	// creature, which is the whole reason Equipment costs more than
	// an Aura for the same stats.
	AttachedAura      float64
	AttachedEquipment float64

	// ManaSource is per untapped mana source; TappedManaSource is
	// what a tapped one is still worth (it untaps next turn).
	ManaSource       float64
	TappedManaSource float64
	FrozenManaSource float64

	// CommanderTax is the penalty per commander cast already made —
	// the {2} surcharge compounds and a seat that has recast its
	// commander three times is genuinely worse off.
	CommanderTax float64

	// Unknown is what a permanent whose face this seat cannot read
	// (a face-down creature) is worth. Deliberately small and
	// positive: it is something rather than nothing.
	Unknown float64

	// OpponentMean and OpponentMax weight the opposition term in
	// Score. Both are subtracted, so the max term is what makes
	// removal land on the table's leader rather than on whoever
	// happens to be first in seat order.
	OpponentMean float64
	OpponentMax  float64

	// ThreatLifeRev is per point of life an opponent is MISSING,
	// for threat ranking only: a player at 12 is a more attractive
	// place to point damage than a player at 38 with the same board.
	ThreatLifeRev float64
	// ThreatHand weights an opponent's hand in the threat ranking —
	// unknown cards are scarier than known ones.
	ThreatHand float64
}

// StartingLife is the Commander starting total the threat ranking
// measures "life missing" against.
const StartingLife = 40

// DefaultWeights is the shipped tuning. The numbers are the usual
// rule-based-AI ratios (Forge's are in the same family) rather than
// anything derived: a 2/2 is worth ~2.9, a land ~1.0, a card in hand
// ~1.2, a point of life ~0.12.
func DefaultWeights() Weights {
	return Weights{
		Life:       0.12,
		LifeDanger: 0.20,
		DangerLife: 10,
		Hand:       1.20,
		Library:    0.01,

		Power:          1.00,
		Toughness:      0.45,
		Keyword:        1.00,
		TappedCreature: 0.85,
		SickCreature:   0.90,
		FrozenCreature: 0.65,
		CantAttack:     0.45,
		CantBlock:      0.70,
		CantActivate:   0.80,

		Permanent:    1.20,
		Planeswalker: 3.00,
		Loyalty:      0.40,

		AttachedAura:      0.10,
		AttachedEquipment: 0.60,

		ManaSource:       1.00,
		TappedManaSource: 0.55,
		FrozenManaSource: 0.25,

		CommanderTax: 1.00,
		Unknown:      1.50,

		OpponentMean: 1.00,
		OpponentMax:  0.50,

		ThreatLifeRev: 0.10,
		ThreatHand:    0.80,
	}
}

// keywordTable is the per-keyword bonus added to a creature's value,
// before the Keyword weight scales it. Keys are the canonical
// lowercase tokens the engine puts on CardView.Abilities (AGENTS.md
// §7). Anything not in the table is worth nothing, which is the
// right default for the long tail of ability words.
var keywordTable = map[string]float64{
	"flying":            1.50,
	"trample":           1.00,
	"deathtouch":        1.60,
	"lifelink":          1.00,
	"first strike":      1.00,
	"double strike":     2.50,
	"vigilance":         0.60,
	"menace":            0.80,
	"reach":             0.40,
	"haste":             0.50,
	"hexproof":          1.60,
	"shroud":            0.60,
	"indestructible":    2.20,
	"protection":        1.20,
	"ward":              1.00,
	"flash":             0.25,
	"skulk":             0.30,
	"fear":              0.80,
	"intimidate":        0.80,
	"horsemanship":      1.20,
	"defender":          -1.50,
	"decayed":           -0.80,
	"cumulative upkeep": -0.60,
}

// keywordBonus sums the table over a card's effective ability list,
// plus its parsed protections.
//
// Protection is counted off CardView.Protection — what the ENGINE's
// one closed-grammar reader made of the tokens (#662) — rather than
// by prefix-matching the ability string. A token the grammar refuses
// is not a protection the engine enforces, and scoring it would price
// a creature above what it actually does. Every quality is worth the
// same here on purpose: how much a protection is worth depends on the
// board, and the threat model this table feeds is deliberately
// board-blind.
//
// Ward still arrives as a prefixed string ("ward {2}") with its cost
// in the parameter and no reader of its own, so it keeps its prefix
// match.
func keywordBonus(c *protocol.CardView) float64 {
	var sum float64
	for _, a := range c.Abilities {
		k := strings.ToLower(strings.TrimSpace(a))
		if v, ok := keywordTable[k]; ok {
			sum += v
			continue
		}
		if strings.HasPrefix(k, "ward") {
			sum += keywordTable["ward"]
		}
	}
	sum += float64(len(c.Protection)) * keywordTable["protection"]
	return sum
}

// SeatEval is the per-seat breakdown Evaluate computes on its way to
// a single number. Exposed because the threat ranking, the attack
// planner and the concede heuristic all want the parts, not the
// total, and recomputing them three times a decision is waste.
type SeatEval struct {
	ID         string
	Seat       int
	Life       int
	Hand       int
	Library    int
	Eliminated bool

	// Creatures is the creature portion of Board; Board is every
	// permanent this seat controls.
	Creatures float64
	Board     float64
	// UntappedMana is the count of untapped permanents that can tap
	// for mana — the bot's read on "can they respond?".
	UntappedMana int
	// CreatureCount and UntappedCreatures are the combat-relevant
	// counts.
	CreatureCount     int
	UntappedCreatures int

	// Strength is the seat's absolute standing: board + life + cards
	// − commander tax.
	Strength float64
}

// isType reports whether a card's effective type line names t.
func isType(c *protocol.CardView, t string) bool {
	return strings.Contains(strings.ToLower(c.TypeLine), t)
}

func isCreature(c *protocol.CardView) bool { return isType(c, "creature") }
func isLand(c *protocol.CardView) bool     { return isType(c, "land") }

// CreatureValue is what one creature on the battlefield is worth.
// Exported because the combat planner trades creatures against each
// other and must use the same scale as the board evaluation.
//
// #727: it applies the RESTRICTION discount. Before that, a creature
// under a Pacifism was worth exactly what it was worth the turn
// before, so the Aura's only effect on the evaluation was the 1.20 it
// cost its own controller as a permanent — casting removal made the
// bot's own score go DOWN. A creature that can neither attack nor
// block is most of a blank card, and this is where the engine's own
// answer to "can it?" (`CardView.Restrictions`, ADR 0045) gets priced.
func (w Weights) CreatureValue(c *protocol.CardView) float64 {
	return w.creatureValue(c, w.restrictionDiscount(c))
}

// creatureValue is CreatureValue with the restriction discount passed
// in, so neutralisedValue can price one creature twice — as it is, and
// as it would be with the restrictions lifted — without a second copy
// of the body.
func (w Weights) creatureValue(c *protocol.CardView, restricted float64) float64 {
	v := w.Power*float64(c.Power) + w.Toughness*float64(c.Toughness) + w.Keyword*keywordBonus(c)
	if v < 0.25 {
		// Even a 0/1 wall is a body; never let a creature price at
		// zero or the blocker logic stops caring whether it dies.
		v = 0.25
	}
	v *= restricted
	if c.Tapped {
		v *= w.TappedCreature
		if WontUntap(c) {
			v *= w.FrozenCreature
		}
	}
	if c.SummoningSick {
		v *= w.SickCreature
	}
	return v
}

// restrictionDiscount is the multiplier for everything the engine
// says this permanent may no longer do (ADR 0045's vocabulary, as the
// stable snake_case tokens on `CardView.Restrictions`). 1 for the
// overwhelming majority of permanents, which are restricted by
// nothing.
//
// The two activation bits collapse into one multiplier: Arrest stops
// mana abilities as well as the rest and Faith's Fetters does not, but
// "its abilities are off" is one loss to the creature's controller
// however many bits the engine needed to say it.
func (w Weights) restrictionDiscount(c *protocol.CardView) float64 {
	if c == nil || len(c.Restrictions) == 0 {
		return 1
	}
	d := 1.0
	activation := false
	for _, r := range c.Restrictions {
		switch r {
		case "cant_attack":
			d *= w.CantAttack
		case "cant_block":
			d *= w.CantBlock
		case "cant_activate", "cant_activate_mana":
			activation = true
		}
	}
	if activation {
		d *= w.CantActivate
	}
	return d
}

// neutralisedValue is how much of a permanent its restrictions have
// taken away: what it would be worth with them lifted, less what it is
// worth now. This is the number a removal Aura is worth to the seat
// that cast it — the debit CreatureValue took off the host's
// controller, credited back on the other side of the ledger.
//
// Zero for a host that is not a creature: the restriction vocabulary's
// other bits gate activations on permanents the evaluation already
// prices flat, so there is nothing to give back.
func (w Weights) neutralisedValue(host *protocol.CardView) float64 {
	if host == nil || !isCreature(host) {
		return 0
	}
	d := w.restrictionDiscount(host)
	if d >= 1 {
		return 0
	}
	return w.creatureValue(host, 1) - w.creatureValue(host, d)
}

// CombatValue is CreatureValue without the tapped and summoning-sick
// discounts. Combat has to compare an attacker — which is tapped the
// moment it is declared — against an untapped blocker, and the board
// evaluation's "a tapped creature is worth less" discount would make
// every attacker look cheap and the bot would never block.
func (w Weights) CombatValue(c *protocol.CardView) float64 {
	v := w.Power*float64(c.Power) + w.Toughness*float64(c.Toughness) + w.Keyword*keywordBonus(c)
	if v < 0.25 {
		v = 0.25
	}
	return v
}

// MarginalLife is what ONE more point of life is worth to a seat at
// the given total: the linear term plus the slope of the low-life
// penalty. At 40 it is small — Commander players do not block to save
// two damage — and it climbs steeply as the total falls, which is
// what turns "never chump" into "chump to survive" without a second
// rule saying so.
func (w Weights) MarginalLife(life int) float64 {
	v := w.Life
	if life < w.DangerLife {
		v += 2 * w.LifeDanger * float64(w.DangerLife-life)
	}
	return v
}

// LifeCostValue is what paying `pay` life costs a seat that is on
// `life` right now, in the units Strength is measured in: the linear
// term plus however much deeper into the danger zone the payment
// takes it. It is exactly the Strength delta Evaluate would compute
// after the payment, which is the point — a life cost has to be
// priced on the same scale as everything it is competing against, or
// the comparison is a fiction.
//
// It is NOT MarginalLife × pay. MarginalLife is the slope at one
// point and the penalty is quadratic, so multiplying it out
// under-prices a big payment from a comfortable total (Griselbrand's
// 7 from 12 life is not seven times the first point) and over-prices
// a small one. Integrating the real curve is the same two
// subtractions, so there is no reason to approximate.
//
// Paying to zero or below is not priced here: that is a loss of the
// game rather than a bad trade, and moves.go refuses it outright.
func (w Weights) LifeCostValue(life, pay int) float64 {
	if pay <= 0 {
		return 0
	}
	return w.Life*float64(pay) + w.lifeDanger(life-pay) - w.lifeDanger(life)
}

// permanentValue prices one permanent on the battlefield.
func (w Weights) permanentValue(c *protocol.CardView) float64 {
	// ADR 0069: a face-down PERMANENT is not unknown — CR 708.2 makes
	// it a public 2/2 colourless creature with no name, and the wire
	// ships that body to every seat. Pricing it as `Unknown` would
	// have a bot ignore a creature it can see, block and kill. What
	// it must not read is the card underneath, and it cannot: the
	// redaction strips the art, the cost and every ability list from
	// a seat that may not look. The Unknown arm is left for cards
	// that really are unreadable — an opponent's hand card surfaced
	// by a prompt, a face-down exile.
	if c.IsFaceDownPermanent() {
		return w.CreatureValue(c)
	}
	if !c.KnownByYou && (c.FaceDown || c.Name == "") {
		return w.Unknown
	}
	switch {
	case isCreature(c):
		return w.CreatureValue(c)
	case isType(c, "planeswalker"):
		return w.Permanent + w.Planeswalker + w.Loyalty*float64(c.Counters["loyalty"])
	case len(c.ManaAbilities) > 0 || isLand(c):
		if c.Tapped {
			if WontUntap(c) {
				return w.FrozenManaSource
			}
			return w.TappedManaSource
		}
		return w.ManaSource
	default:
		return w.Permanent
	}
}

// AttachRole is what an attached permanent is DOING to the board, and
// it is the one classification #727 asked for: four answers, derived
// from the effect the attachment's statics have ALREADY had on its
// host, never from the card's name.
//
// Reading the host rather than the card is the whole trick. The
// heuristic holds a protocol.GameView and may not hold a *game.Game
// (ADR 0033 §3), so it cannot open a CardDef and walk its
// StaticAbility list — but it does not have to. Every one of those
// statics has already run: the layer pass has put the +2/+2 on the
// host's Power and Toughness, the restriction bits on its
// Restrictions, and the layer-2 control change on its Controller. So
// what an attachment DOES is legible from what its host now IS, which
// is also the only reading that stays correct when a card the catalog
// has never heard of does the same thing.
type AttachRole int

const (
	// AttachNone — not attached to anything the evaluation can see:
	// an Equipment nobody has equipped, or one whose host has left and
	// whose CR 704.5n unattach has not run yet. Priced as its own
	// permanent, which is what it is.
	AttachNone AttachRole = iota

	// AttachBuff — an Equipment or a +N/+N Aura. The value RIDES THE
	// HOST: the host's Power, Toughness and Abilities on the wire are
	// post-layer, so the boost is already counted there, and pricing
	// the attachment at w.Permanent as well was the double count
	// ADR 0036's hand-off named. What is left on the attachment's own
	// line is its reattach / utility value and nothing else.
	AttachBuff

	// AttachRestriction — a Pacifism or an Arrest on a permanent its
	// controller does not control. The value is the HOST it
	// neutralises: CreatureValue has already debited the host's
	// controller for the restriction, and this is the matching credit
	// to the seat that cast the Aura. Without it the bot can see the
	// opponent get worse but not that IT did that, and its own removal
	// reads as a permanent it paid 1.20 for.
	AttachRestriction

	// AttachControl — a Control Magic or a Mind Control. Nothing
	// extra: the layer-2 effect has already moved the creature onto
	// this seat's side of the ledger, where the ordinary creature pass
	// counts it. Counting the Aura too would pay for the theft twice.
	AttachControl

	// AttachCurse — attached to a PLAYER (Curse of Opulence). There is
	// no host permanent to carry anything, so it keeps its own line
	// and is priced as its own permanent, exactly as it was.
	AttachCurse
)

// neutralisingRestriction reports whether a restriction token makes
// its permanent WORSE for the seat that controls it. The one that does
// not is "cant_be_blocked", which is why the set is spelled out rather
// than taken as "Restrictions is non-empty".
func neutralisingRestriction(r string) bool {
	switch r {
	case "cant_attack", "cant_block", "cant_activate", "cant_activate_mana":
		return true
	}
	return false
}

// isNeutralised reports whether anything has taken this permanent's
// declaration- or activation-time options away.
func isNeutralised(c *protocol.CardView) bool {
	if c == nil {
		return false
	}
	for _, r := range c.Restrictions {
		if neutralisingRestriction(r) {
			return true
		}
	}
	return false
}

// AttachmentRole classifies one attached permanent against its host.
// `host` is the battlefield card `c.AttachedTo` names, or nil when the
// attachment names a player or the host is not on the battlefield.
//
// The arms are in order of certainty:
//
//   - a player host is a curse, and nothing else can be;
//   - a host whose CONTROLLER is the attachment's controller while its
//     OWNER is somebody else has been stolen, and the layer-2 effect
//     that stole it is the thing attached to it;
//   - a host under a neutralising restriction whose controller is not
//     the attachment's is being answered, not helped;
//   - everything else is a buff.
//
// The two soft edges are both cheap. A buff Aura on a creature this
// seat stole with something ELSE reads as control and is priced at 0
// instead of AttachedAura — a tenth of a point. A second restriction
// Aura on an already-pacified creature claims the same neutralised
// value as the first; the board where that happens has two seats
// spending two cards to answer one creature, and over-rating their
// answers is not a decision anyone is worried about.
func AttachmentRole(c, host *protocol.CardView) AttachRole {
	if c == nil || c.AttachedTo == nil || c.AttachedTo.ID == "" {
		return AttachNone
	}
	if c.AttachedTo.Kind == "player" {
		return AttachCurse
	}
	if host == nil {
		return AttachNone
	}
	if host.Controller == c.Controller && host.Owner != "" && host.Owner != c.Controller {
		return AttachControl
	}
	if host.Controller != c.Controller && isNeutralised(host) {
		return AttachRestriction
	}
	return AttachBuff
}

// isEquipment reports whether a permanent is one of the attachment
// types that SURVIVES its host — Equipment (CR 301.5) and Fortification
// (CR 301.6). Both unattach rather than die when the attachment becomes
// illegal (CR 704.5n), which is exactly why they keep a bigger residual
// than an Aura does.
func isEquipment(c *protocol.CardView) bool {
	return isType(c, "equipment") || isType(c, "fortification")
}

// attachIndex resolves the battlefield's attachment relation once per
// evaluation: instance ID to card, so an attachment can find its host.
// Built from the same slice the evaluation walks, so it costs one extra
// pass over a board of a few dozen cards.
type attachIndex map[string]*protocol.CardView

func newAttachIndex(cards []protocol.CardView) attachIndex {
	ix := make(attachIndex, len(cards))
	for i := range cards {
		ix[cards[i].InstanceID] = &cards[i]
	}
	return ix
}

// hostOf is the permanent `c` is attached to, or nil — for an
// unattached permanent, a curse on a player, and a dangling attachment
// whose host has already left.
func (ix attachIndex) hostOf(c *protocol.CardView) *protocol.CardView {
	if c == nil || c.AttachedTo == nil || c.AttachedTo.Kind != "card" || c.AttachedTo.ID == "" {
		return nil
	}
	return ix[c.AttachedTo.ID]
}

// boardValue prices one permanent for the seat that controls it with
// the attachment relation resolved: permanentValue for anything that is
// not attached, and the role's price for anything that is. This is the
// ONE place #727's split lives; every other caller asks it.
func (w Weights) boardValue(c *protocol.CardView, ix attachIndex) float64 {
	host := ix.hostOf(c)
	switch AttachmentRole(c, host) {
	case AttachControl:
		return 0
	case AttachRestriction:
		return w.neutralisedValue(host)
	case AttachBuff:
		if isEquipment(c) {
			return w.AttachedEquipment
		}
		return w.AttachedAura
	}
	return w.permanentValue(c)
}

// WontUntap reads the public projection for the controller's next untap
// step. A marker naming another player doesn't freeze this resource for
// its controller, and an untapped permanent is still available now.
func WontUntap(c *protocol.CardView) bool {
	if c == nil || !c.Tapped || c.NoUntap == nil {
		return false
	}
	if c.NoUntap.Static {
		return true
	}
	for _, id := range c.NoUntap.Next {
		if id == c.Controller {
			return true
		}
	}
	return false
}

// producesMana reports whether a permanent can be tapped for mana
// right now.
func producesMana(c *protocol.CardView) bool {
	if c.Tapped {
		return false
	}
	if len(c.ManaAbilities) > 0 {
		return !(isCreature(c) && c.SummoningSick)
	}
	return isLand(c)
}

// Evaluate breaks the view down per seat: one index pass and one
// scoring pass over the battlefield, one pass over the seats.
//
// The index pass is #727. An attachment is priced against its host, so
// the host has to be findable before anything is added up — and the
// host may appear after the attachment in the battlefield slice, so a
// single pass cannot do it.
func (w Weights) Evaluate(v protocol.GameView) map[string]*SeatEval {
	out := make(map[string]*SeatEval, len(v.Seats))
	for i := range v.Seats {
		s := &v.Seats[i]
		out[s.ID] = &SeatEval{
			ID:         s.ID,
			Seat:       s.Seat,
			Life:       s.Life,
			Hand:       s.Hand.Count,
			Library:    s.Library.Count,
			Eliminated: s.Eliminated,
		}
	}
	ix := newAttachIndex(v.Battlefield.Cards)
	for i := range v.Battlefield.Cards {
		c := &v.Battlefield.Cards[i]
		e := out[c.Controller]
		if e == nil {
			continue
		}
		e.Board += w.boardValue(c, ix)
		if isCreature(c) {
			e.Creatures += w.CreatureValue(c)
			e.CreatureCount++
			if !c.Tapped {
				e.UntappedCreatures++
			}
		}
		if producesMana(c) {
			e.UntappedMana++
		}
	}
	for i := range v.Seats {
		s := &v.Seats[i]
		e := out[s.ID]
		if e.Eliminated {
			e.Strength = eliminatedStrength
			continue
		}
		var tax float64
		for _, n := range s.CommanderCasts {
			tax += float64(n)
		}
		e.Strength = e.Board +
			w.Life*float64(e.Life) +
			w.Hand*float64(e.Hand) +
			w.Library*float64(e.Library) -
			w.CommanderTax*tax -
			w.lifeDanger(e.Life)
	}
	return out
}

// eliminatedStrength is the standing of a seat that is out of the
// game: far below anything a live seat can reach, so "eliminate that
// player" always dominates "shrink their board".
const eliminatedStrength = -1000.0

func (w Weights) lifeDanger(life int) float64 {
	if life >= w.DangerLife {
		return 0
	}
	d := float64(w.DangerLife - life)
	return w.LifeDanger * d * d
}

// Score is the positional evaluation of the whole table from one
// seat's point of view: my standing, less the table's. The mean term
// is what makes "everyone else is bigger than me" a bad position
// even when no single opponent is; the max term is what makes the
// LEADER the right place to point removal.
//
// perspective is a seat's player-UUID string, as it appears in
// PlayerView.ID. An unknown perspective scores 0.
func Score(v protocol.GameView, perspective string) float64 {
	return DefaultWeights().Score(v, perspective)
}

// Score is the Weights-tunable form of the package-level Score.
func (w Weights) Score(v protocol.GameView, perspective string) float64 {
	return w.ScoreEval(w.Evaluate(v), perspective)
}

// ScoreEval is Score over an already-computed breakdown.
func (w Weights) ScoreEval(evals map[string]*SeatEval, perspective string) float64 {
	me := evals[perspective]
	if me == nil {
		return 0
	}
	var sum, max float64
	n := 0
	for id, e := range evals {
		if id == perspective || e.Eliminated {
			continue
		}
		sum += e.Strength
		if n == 0 || e.Strength > max {
			max = e.Strength
		}
		n++
	}
	if n == 0 {
		return me.Strength
	}
	return me.Strength - w.OpponentMean*(sum/float64(n)) - w.OpponentMax*max
}

// Threat ranks an opponent as a place to point damage and removal:
// their board, their hidden resources, and how close they already
// are to dying. It is deliberately NOT Strength — a player at 8 life
// with an empty board is a low-strength seat and a high-priority
// attack target, and the bot has to be able to tell those apart.
func (w Weights) Threat(e *SeatEval) float64 {
	if e == nil || e.Eliminated {
		return eliminatedStrength
	}
	missing := StartingLife - e.Life
	if missing < 0 {
		missing = 0
	}
	return e.Board + w.ThreatHand*float64(e.Hand) + w.ThreatLifeRev*float64(missing)
}
