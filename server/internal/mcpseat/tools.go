package mcpseat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/boardtext"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// The ten tools' inputs (§3). Plain structs of ours: transport.go hands
// them to the SDK, which infers each tool's schema from the json and
// jsonschema tags. A field with omitempty is optional.

// JoinInput is `join`'s input.
type JoinInput struct {
	InviteURL   string     `json:"invite_url" jsonschema:"the table's invite link (…/#/games/<id>/join?t=…) or an admin's seat-reclaim link"`
	DisplayName string     `json:"display_name,omitempty" jsonschema:"the seat's name at the table (default Agent)"`
	Deck        *DeckInput `json:"deck,omitempty" jsonschema:"optional deck to install: {id} for a pre-built deck or {list} for a decklist"`
}

// SetDeckInput is `set_deck`'s input.
type SetDeckInput struct {
	Deck DeckInput `json:"deck" jsonschema:"{id} for a pre-built deck or {list} for a decklist"`
}

// WaitInput is `wait_for_decision`'s input.
type WaitInput struct {
	TimeoutS  int    `json:"timeout_s,omitempty" jsonschema:"seconds to wait, 1 to 50 (default 25)"`
	PassUntil string `json:"pass_until,omitempty" jsonschema:"none (default) or my_turn_or_stack: also pass priority on other players' turns while the stack is empty"`
}

// GetStateInput is `get_state`'s input.
type GetStateInput struct {
	Detail string `json:"detail,omitempty" jsonschema:"compact (default) or full"`
}

// LegalMovesInput is `legal_moves`'s input.
type LegalMovesInput struct {
	Card   string `json:"card,omitempty" jsonschema:"optional card instance id: list that card's moves with the enumerator's caps lifted"`
	Choice string `json:"choice,omitempty" jsonschema:"optional pending choice id, or cleanup_discard: list that prompt's answers with the caps lifted"`
}

// CardInput is `card`'s input.
type CardInput struct {
	Ref string `json:"ref" jsonschema:"a card instance id you can see, or the name of a card on the table"`
}

// ActInput is `act`'s input.
type ActInput struct {
	Window string `json:"window" jsonschema:"the window token from wait_for_decision"`
	Move   int    `json:"move" jsonschema:"the number of the move in that window's list"`
	Value  any    `json:"value,omitempty" jsonschema:"only for a move marked open: a card name, or a number for X"`
}

// SayInput is `say`'s input.
type SayInput struct {
	Text string `json:"text" jsonschema:"1 to 500 characters of chat"`
}

// ConcedeInput is `concede`'s input.
type ConcedeInput struct {
	Confirm bool `json:"confirm" jsonschema:"must be true"`
}

// LeaveInput is `leave`'s input: none.
type LeaveInput struct{}

// --- join ---------------------------------------------------------------

// Join claims a guest seat by invite link, or reattaches a seat this
// binary already holds, and connects to the table.
func (s *Seat) Join(ctx context.Context, in JoinInput) (Result, error) {
	s.opMu.Lock()
	defer s.opMu.Unlock()

	inv, err := parseInvite(in.InviteURL)
	if err != nil {
		return errorResult("%v", err), nil
	}
	u, _ := url.Parse(inv.Origin)
	if !s.cfg.Origins.Allows(u) {
		return errorResult("this seat only talks to the servers its owner allowed (%s). %s is not one of them; the owner sets --allow-origin.",
			s.cfg.Origins, inv.Origin), nil
	}

	s.mu.Lock()
	held := s.sess
	client := s.client
	s.mu.Unlock()
	if held != nil {
		if held.GameID == inv.GameID && held.Origin == inv.Origin {
			return s.joinedText(true, nil), nil
		}
		return errorResult("this binary already holds a seat at game %s; one binary holds one seat. Call leave first.", held.GameID), nil
	}

	name := strings.TrimSpace(in.DisplayName)
	if name == "" {
		name = defaultName
	}
	name = cut(name, maxNameLen)

	sess, resumed, err := s.claim(ctx, inv, name, agentClientName(client))
	if err != nil {
		return errorResult("could not take the seat: %v", err), nil
	}
	s.sec.add(sess.Token)
	if !resumed {
		if err := s.store.save(sess); err != nil {
			s.log.Warn("could not save the session; a restart will not reattach", "err", err)
		}
	}
	if err := s.connect(sess); err != nil {
		return errorResult("took the seat but could not connect: %v", err), nil
	}
	s.awaitFirstSnapshot(ctx)

	var deckNote []string
	if in.Deck != nil {
		r := s.installDeck(ctx, *in.Deck)
		deckNote = append(deckNote, r.Text)
	}
	return s.joinedText(resumed, deckNote), nil
}

// claim finds a saved session for the table and checks it still works, or
// joins (or redeems a reclaim ticket) for a new one.
func (s *Seat) claim(ctx context.Context, inv invite, name, client string) (*savedSession, bool, error) {
	if inv.Kind == inviteJoin {
		saved, err := s.store.load(inv.Origin, inv.GameID)
		if err != nil {
			return nil, false, fmt.Errorf("the saved session cannot be used: %w", err)
		}
		if saved != nil && s.now().Before(saved.ExpiresAt) {
			s.sec.add(saved.Token)
			p, err := s.api.me(ctx, inv.Origin, saved.Token)
			var ae *apiError
			switch {
			case err == nil && p.GameID == inv.GameID && p.PlayerID == saved.PlayerID:
				return saved, true, nil
			case err == nil, errors.As(err, &ae) && ae.Status == http.StatusUnauthorized:
				_ = s.store.remove(inv.Origin, inv.GameID)
			default:
				return nil, false, fmt.Errorf("could not check the saved session: %w", err)
			}
		} else if saved != nil {
			_ = s.store.remove(inv.Origin, inv.GameID)
		}
	}
	var ans *sessionAnswer
	var err error
	if inv.Kind == inviteReclaim {
		ans, err = s.api.reclaim(ctx, inv)
	} else {
		ans, err = s.api.join(ctx, inv, name, client)
	}
	if err != nil {
		return nil, false, err
	}
	return &savedSession{
		Origin:    inv.Origin,
		GameID:    inv.GameID,
		PlayerID:  ans.playerID(),
		Token:     ans.Token,
		ExpiresAt: ans.ExpiresAt,
	}, false, nil
}

// connect opens the socket and starts the autopilot.
func (s *Seat) connect(sess *savedSession) error {
	wsURL, err := wsURLFor(sess.Origin, sess.GameID.String(), sess.PlayerID.String())
	if err != nil {
		return err
	}
	conn := newWSConn(wsURL, sess.Origin, sess.Token, s.cfg.Dialer, s.api, s, s.log)
	stop, done := make(chan struct{}), make(chan struct{})
	s.mu.Lock()
	s.sess = sess
	s.conn = conn
	s.endReason = ""
	s.poke = make(chan struct{}, 1)
	s.stopAuto, s.autoDone = stop, done
	s.mu.Unlock()
	go s.autopilot(stop, done)
	conn.start()
	return nil
}

// awaitFirstSnapshot waits briefly for the table to arrive.
func (s *Seat) awaitFirstSnapshot(ctx context.Context) {
	deadline := time.NewTimer(firstSnapshotWait)
	defer deadline.Stop()
	for {
		s.mu.Lock()
		ok, ended, ch := s.haveView, s.endReason != "", s.changed
		s.mu.Unlock()
		if ok || ended {
			return
		}
		select {
		case <-ch:
		case <-deadline.C:
			return
		case <-ctx.Done():
			return
		}
	}
}

// joinedText is join's answer: the seat and the table.
func (s *Seat) joinedText(resumed bool, extra []string) Result {
	s.mu.Lock()
	sess := s.sess
	view := s.view
	ended := s.endReason
	s.mu.Unlock()
	lines := []string{
		"joined: true",
		fmt.Sprintf("resumed: %t", resumed),
		"game_id: " + sess.GameID.String(),
		"player_id: " + sess.PlayerID.String(),
	}
	if ended != "" {
		lines = append(lines, "connection: "+ended)
	}
	if view == nil {
		lines = append(lines, "table: not received yet (call wait_for_decision)")
		return textResult(append(lines, extra...)...)
	}
	if self := mySeat(view, sess.PlayerID.String()); self != nil {
		lines = append(lines, fmt.Sprintf("seat: %d", self.Seat))
	}
	lines = append(lines, "table state: "+view.State)
	lines = append(lines, seatList(view, sess.PlayerID.String())...)
	if view.State == "lobby" {
		if self := mySeat(view, sess.PlayerID.String()); self != nil && !self.DeckImported {
			lines = append(lines, "your deck: none yet. Call set_deck before the host starts the game.")
		}
	}
	lines = append(lines, extra...)
	lines = append(lines, "", "Next: loop wait_for_decision → act until status is game_over. "+untrustedNote)
	return textResult(lines...)
}

// seatList is one line per seat, names wrapped as table text.
func seatList(v *protocol.GameView, me string) []string {
	out := []string{"seats:"}
	for i := range v.Seats {
		p := &v.Seats[i]
		line := fmt.Sprintf("  seat %d: %s", p.Seat, untrusted(p.Name, maxNameLen))
		var tags []string
		if p.ID == me {
			tags = append(tags, "you")
		}
		if p.IsBot {
			tags = append(tags, "bot")
		}
		if p.IsHost {
			tags = append(tags, "host")
		}
		if v.State == "lobby" {
			if p.DeckImported {
				tags = append(tags, "deck ready")
			} else {
				tags = append(tags, "no deck yet")
			}
		}
		if p.Eliminated {
			tags = append(tags, "eliminated")
		}
		if len(tags) > 0 {
			line += " (" + strings.Join(tags, ", ") + ")"
		}
		out = append(out, line)
	}
	return out
}

// --- set_deck -----------------------------------------------------------

// SetDeck installs a deck on the seat, before the game starts.
func (s *Seat) SetDeck(ctx context.Context, in SetDeckInput) (Result, error) {
	s.opMu.Lock()
	defer s.opMu.Unlock()
	return s.installDeck(ctx, in.Deck), nil
}

func (s *Seat) installDeck(ctx context.Context, d DeckInput) Result {
	d.ID, d.List = strings.TrimSpace(d.ID), strings.TrimSpace(d.List)
	if (d.ID == "") == (d.List == "") {
		return errorResult("deck takes exactly one of id (a pre-built deck) or list (a decklist)")
	}
	s.mu.Lock()
	sess, view := s.sess, s.view
	s.mu.Unlock()
	if sess == nil {
		return errorResult("not seated: call join first")
	}
	if view != nil && view.State != "lobby" {
		return errorResult("the game has started; a deck can only be set before the start")
	}
	ans, err := s.api.setDeck(ctx, sess.Origin, sess.Token, sess.GameID, sess.PlayerID, d)
	if err != nil {
		return errorResult("the server refused the deck: %v", err)
	}
	lines := []string{fmt.Sprintf("deck installed: %s (%d cards)", untrusted(ans.DeckName, 80), ans.CardCount)}
	if len(ans.Commanders) > 0 {
		lines = append(lines, "commanders: "+strings.Join(ans.Commanders, ", "))
	}
	for _, w := range ans.Warnings {
		lines = append(lines, "warning: "+w.Message)
	}
	if len(ans.Unimplemented) > 0 {
		lines = append(lines, "the engine does not run these cards' text (they resolve to nothing; say in chat what they should do): "+
			strings.Join(ans.Unimplemented, ", "))
	}
	return textResult(lines...)
}

// --- wait_for_decision ----------------------------------------------------

// WaitForDecision blocks until the seat has a real choice, the game ends,
// or the timeout passes.
func (s *Seat) WaitForDecision(ctx context.Context, in WaitInput) (Result, error) {
	timeout := in.TimeoutS
	if timeout <= 0 {
		timeout = defaultWait
	}
	if timeout > maxWait {
		timeout = maxWait
	}
	switch in.PassUntil {
	case "", "none":
	case "my_turn_or_stack":
	default:
		return errorResult("pass_until is none or my_turn_or_stack"), nil
	}
	s.mu.Lock()
	if s.sess == nil {
		s.mu.Unlock()
		return errorResult("not seated: call join first"), nil
	}
	s.passUntil = in.PassUntil == "my_turn_or_stack"
	s.mu.Unlock()
	s.pokeAutopilot()

	deadline := time.NewTimer(time.Duration(timeout) * time.Second)
	defer deadline.Stop()
	for {
		s.mu.Lock()
		r, done := s.statusLocked()
		ch := s.changed
		s.mu.Unlock()
		if done {
			return r, nil
		}
		select {
		case <-ch:
		case <-deadline.C:
			s.mu.Lock()
			r := s.notReadyLocked()
			s.mu.Unlock()
			return r, nil
		case <-ctx.Done():
			return textResult("status: waiting", "(cancelled)"), nil
		}
	}
}

// statusLocked answers wait_for_decision if there is anything to say now.
func (s *Seat) statusLocked() (Result, bool) {
	if s.endReason != "" {
		return errorResult("status: disconnected\nThe connection to the table ended: %s", s.endReason), true
	}
	if !s.haveView {
		return Result{}, false
	}
	v := s.view
	if gameOver(v) {
		return s.gameOverLocked(), true
	}
	if v.State != "active" {
		return Result{}, false
	}
	if self := s.selfLocked(); self != nil && self.Eliminated {
		return textResult("status: eliminated",
			"Your seat is out of the game; the others play on. Call leave to disconnect, or keep calling wait_for_decision to watch for game_over."), true
	}
	if w := s.win; w != nil && w.key == s.key && w.state == winDecision {
		return s.decisionTextLocked(w), true
	}
	return Result{}, false
}

// notReadyLocked is the answer when the timeout passes with nothing to
// decide.
func (s *Seat) notReadyLocked() Result {
	if !s.haveView {
		return textResult("status: waiting", "Still connecting to the table. Call wait_for_decision again.")
	}
	if s.view.State == "lobby" {
		lines := []string{"status: not_started", "The host has not started the game yet. Call wait_for_decision again."}
		lines = append(lines, seatList(s.view, s.sess.PlayerID.String())...)
		return textResult(lines...)
	}
	line := "status: waiting"
	if !s.isConnected {
		line += " (reconnecting to the table)"
	}
	return textResult(line, "Nothing for you to decide yet. Call wait_for_decision again.")
}

// gameOverLocked is the game_over answer, with §9's report.
func (s *Seat) gameOverLocked() Result {
	v := s.view
	lines := []string{"status: game_over"}
	if o := v.Outcome; o != nil {
		switch {
		case o.Winner != "":
			name := boardtext.SeatLabelByID(safeView(v), o.Winner, s.sess.PlayerID.String())
			lines = append(lines, fmt.Sprintf("outcome: %s wins (%s)", name, o.Cause))
		case o.Kind != "":
			lines = append(lines, fmt.Sprintf("outcome: %s (%s)", o.Kind, o.Cause))
		}
	} else {
		lines = append(lines, "outcome: the table ended with no result")
	}
	lines = append(lines, "", s.stats.report(), "Call leave to disconnect.")
	return textResult(lines...)
}

// decisionTextLocked is the decision answer: the window, its kind, the
// compact view and the numbered moves, with what happened since the last
// decision. It resets the since-last-decision trackers.
func (s *Seat) decisionTextLocked(w *window) Result {
	v := s.view
	me := s.sess.PlayerID.String()
	nw := newNameWrapper(v)
	var logLines []string
	var maxSeq uint64
	for _, e := range v.Log {
		if e.Seq > s.logSeqShown {
			logLines = append(logLines, logLine(e, nw))
		}
		if e.Seq > maxSeq {
			maxSeq = e.Seq
		}
	}
	if len(logLines) > sinceLogLines {
		logLines = logLines[len(logLines)-sinceLogLines:]
	}
	var head []string
	head = append(head, "status: decision", "window: "+w.token, "kind: "+w.kind)
	if self := mySeat(v, me); self != nil {
		head = append(head, fmt.Sprintf("you: %s (seat %d)", untrusted(self.Name, maxNameLen), self.Seat))
	}
	if v.LoopNotice != nil {
		head = append(head, fmt.Sprintf("LOOP NOTICE (CR 732): %s has resolved %d times this turn. Nothing passes automatically until a player decides.",
			nw.apply(v.LoopNotice.Label), v.LoopNotice.Count))
	}
	if sum := absorbedSummary(s.autoSince); sum != "" {
		head = append(head, sum)
	}
	board := compactBoard(v, me, logLines, s.chatSince)
	moves := renderMoves(v, w, "")
	s.stats.noteMovesShown(len(w.moves), len(moves))

	s.autoSince = map[string]int{}
	s.chatSince = nil
	if maxSeq > s.logSeqShown {
		s.logSeqShown = maxSeq
	}
	return textResult(strings.Join(head, "\n"), "", board, "", moves)
}

// --- get_state ----------------------------------------------------------

// GetState is the view at the compact or the full budget.
func (s *Seat) GetState(_ context.Context, in GetStateInput) (Result, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.sess == nil {
		return errorResult("not seated: call join first"), nil
	}
	if !s.haveView {
		return errorResult("no table yet: still connecting"), nil
	}
	me := s.sess.PlayerID.String()
	header := fmt.Sprintf("table state: %s", s.view.State)
	if w := s.win; w != nil && w.key == s.key && w.state == winDecision {
		header += fmt.Sprintf("; open decision: window %s (%s)", w.token, w.kind)
	}
	switch in.Detail {
	case "", "compact":
		return textResult(header, untrustedNote, "", compactBoard(s.view, me, nil, nil)), nil
	case "full":
		return textResult(header, "", fullBoard(s.view, me)), nil
	default:
		return errorResult("detail is compact or full"), nil
	}
}

// --- legal_moves ----------------------------------------------------------

// LegalMoves is the current window's full numbered list, or one card's or
// one prompt's moves with the enumerator's caps lifted (§6.2).
func (s *Seat) LegalMoves(ctx context.Context, in LegalMovesInput) (Result, error) {
	req := moveRequest{Source: strings.TrimSpace(in.Card), Choice: strings.TrimSpace(in.Choice)}
	s.mu.Lock()
	if s.sess == nil {
		s.mu.Unlock()
		return errorResult("not seated: call join first"), nil
	}
	w := s.win
	if w == nil || w.key != s.key || w.state != winDecision {
		s.mu.Unlock()
		return errorResult("no decision is open for you right now: call wait_for_decision"), nil
	}
	view := s.view
	if req.Source == "" && req.Choice == "" {
		text := renderMoves(view, w, "")
		s.stats.noteMovesShown(len(w.moves), len(text))
		s.mu.Unlock()
		return textResult("window: "+w.token, text), nil
	}
	if req.Source != "" && req.Choice != "" {
		s.mu.Unlock()
		return errorResult("name a card or a choice, not both"), nil
	}
	if req.Source != "" {
		if c, _ := findCard(view, req.Source); c == nil {
			s.mu.Unlock()
			return errorResult("no card with instance id %q is visible to you", req.Source), nil
		}
	}
	s.mu.Unlock()

	rep, err := s.requestMoves(ctx, req)
	if errors.Is(err, errNoDecision) {
		s.mu.Lock()
		defer s.mu.Unlock()
		return s.staleLocked(), nil
	}
	if err != nil {
		s.mu.Lock()
		text := renderMoves(view, w, req.Source)
		s.mu.Unlock()
		return textResult("could not expand it: "+err.Error(), "", "window: "+w.token, text), nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.win != w || rep.Generation != w.key.gen || rep.Seq != w.key.seq {
		return s.staleLocked(), nil
	}
	added := mergeMoves(w, rep, req)
	text := renderMoves(view, w, req.Source)
	s.stats.noteMovesShown(len(w.moves), len(text))
	note := fmt.Sprintf("%d moves added to this window's list.", added)
	if len(rep.Moves) == 0 {
		note = "The server lists no moves for it right now."
	}
	return textResult("window: "+w.token, note, text), nil
}

// mergeMoves appends a card's expanded moves to the window, so every
// index still names one move of one window, and replaces that card's
// cut report with the expanded request's.
func mergeMoves(w *window, rep *protocol.LegalMovesPayload, req moveRequest) int {
	have := map[string]bool{}
	for _, m := range w.moves {
		have[moveKey(m)] = true
	}
	added := 0
	for _, m := range rep.Moves {
		if k := moveKey(m); !have[k] {
			have[k] = true
			w.moves = append(w.moves, m)
			added++
		}
	}
	cuts := w.cuts[:0:0]
	for _, c := range w.cuts {
		if !(req.Source != "" && c.Source == req.Source) && !(req.Choice != "" && c.Choice == req.Choice) {
			cuts = append(cuts, c)
		}
	}
	cuts = append(cuts, rep.Truncated...)
	w.cuts = cuts
	return added
}

func moveKey(m legal.Move) string {
	return m.Type + "|" + m.Player.String() + "|" + string(m.Params)
}

// staleLocked is the answer to an index into a window that has closed.
func (s *Seat) staleLocked() Result {
	if w := s.win; w != nil && w.key == s.key && w.state == winDecision {
		r := s.decisionTextLocked(w)
		return textResult("status: stale — the board moved; this is the new window. Nothing was sent.", "", r.Text)
	}
	return textResult("status: stale — the board moved and no decision is open now. Nothing was sent. Call wait_for_decision.")
}

// --- card ---------------------------------------------------------------

// Card shows one card's printed text, from the server's card route.
func (s *Seat) Card(ctx context.Context, in CardInput) (Result, error) {
	s.mu.Lock()
	sess, view := s.sess, s.view
	s.mu.Unlock()
	if sess == nil || view == nil {
		return errorResult("not seated: call join first"), nil
	}
	ref := strings.TrimSpace(in.Ref)
	c, zone := findCard(view, ref)
	if c == nil {
		eachCard(view, func(cv *protocol.CardView, z string) bool {
			if cv.Name != "" && strings.EqualFold(cv.Name, ref) && (!cv.FaceDown || cv.FaceVisible) {
				c, zone = cv, z
				return false
			}
			return true
		})
	}
	if c == nil {
		return errorResult("no card %q is visible to you", cut(ref, 80)), nil
	}
	if c.IsFaceDownPermanent() || (c.FaceDown && !c.FaceVisible) || c.Name == "" {
		return textResult(boardtext.CardName(c) + " (" + zone + "): you may not see its face."), nil
	}
	var lines []string
	lines = append(lines, fmt.Sprintf("%s (%s) [%s]", c.Name, zone, c.InstanceID))
	if c.Unimplemented {
		lines = append(lines, boardtext.UnimplementedNote+". If it should do something, say so in chat for a person to resolve.")
	}
	meta, err := s.cardMeta(ctx, sess, c.ScryfallID)
	switch {
	case c.ScryfallID == "":
		lines = append(lines, cardFromView(c)...)
	case err != nil:
		lines = append(lines, cardFromView(c)...)
		lines = append(lines, "(oracle text unavailable: "+err.Error()+")")
	default:
		lines = append(lines, cardFromMeta(meta)...)
	}
	return Result{Text: hardCut(strings.Join(lines, "\n"), budgetCard)}, nil
}

func cardFromView(c *protocol.CardView) []string {
	var out []string
	if c.TypeLine != "" {
		out = append(out, "type: "+c.TypeLine)
	}
	if c.ManaCost != "" {
		out = append(out, "cost: "+c.ManaCost)
	}
	if strings.Contains(strings.ToLower(c.TypeLine), "creature") {
		out = append(out, fmt.Sprintf("power/toughness now: %d/%d", c.Power, c.Toughness))
	}
	return out
}

func cardFromMeta(m *cardMeta) []string {
	var out []string
	face := func(name, typ, cost, text, p, t, loy string) {
		if name != "" {
			out = append(out, "— "+name)
		}
		if typ != "" {
			out = append(out, "type: "+typ)
		}
		if cost != "" {
			out = append(out, "cost: "+cost)
		}
		if p != "" || t != "" {
			out = append(out, "power/toughness: "+p+"/"+t)
		}
		if loy != "" {
			out = append(out, "loyalty: "+loy)
		}
		if text != "" {
			out = append(out, "text: "+text)
		}
	}
	if m.OracleText == "" && len(m.CardFaces) > 0 {
		for _, f := range m.CardFaces {
			face(f.Name, f.TypeLine, f.ManaCost, f.OracleText, f.Power, f.Toughness, f.Loyalty)
		}
		return out
	}
	face("", m.TypeLine, m.ManaCost, m.OracleText, m.Power, m.Toughness, m.Loyalty)
	return out
}

// cardMeta fetches a card's metadata once per id, at most 5 a second.
func (s *Seat) cardMeta(ctx context.Context, sess *savedSession, id string) (*cardMeta, error) {
	if id == "" {
		return nil, errors.New("no printed card behind it")
	}
	s.mu.Lock()
	m, ok := s.cardCache[id]
	s.mu.Unlock()
	if ok {
		return m, nil
	}
	if err := s.cards.wait(ctx); err != nil {
		return nil, err
	}
	m, err := s.api.card(ctx, sess.Origin, sess.Token, id)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	s.cardCache[id] = m
	s.mu.Unlock()
	return m, nil
}

// --- act ----------------------------------------------------------------

// Act sends one move of the current window, and reports the server's ack
// or error (§3, §6.4). Nothing is inferred from other seats' snapshots.
func (s *Seat) Act(ctx context.Context, in ActInput) (Result, error) {
	s.mu.Lock()
	if s.sess == nil {
		s.mu.Unlock()
		return errorResult("not seated: call join first"), nil
	}
	w := s.win
	if w == nil || w.key != s.key || w.token != in.Window || w.state != winDecision {
		r := s.staleLocked()
		s.mu.Unlock()
		return r, nil
	}
	if in.Move < 0 || in.Move >= len(w.moves) {
		s.mu.Unlock()
		return errorResult("move %d is not in this window's list (0 to %d)", in.Move, len(w.moves)-1), nil
	}
	m := w.moves[in.Move]
	if w.rejections >= maxRejections && !m.AlwaysLegal {
		var safe []string
		for i, mv := range w.moves {
			if mv.AlwaysLegal {
				safe = append(safe, fmt.Sprintf("%d (%s)", i, mv.Label))
			}
		}
		s.mu.Unlock()
		return errorResult("three moves were refused in this window; only an always-legal move is accepted now: %s", strings.Join(safe, ", ")), nil
	}
	s.mu.Unlock()

	params, err := applyValue(m, in.Value)
	if err != nil {
		return errorResult("%v", err), nil
	}
	if err := s.actions.wait(ctx); err != nil {
		return textResult("status: cancelled", "Nothing was sent."), nil
	}
	ch, err := s.sendAction(m, params)
	if err != nil {
		return errorResult("status: not_sent\n%v", err), nil
	}
	s.mu.Lock()
	s.stats.acts++
	s.mu.Unlock()

	wait, legacy := s.ackWait()
	t := time.NewTimer(wait)
	defer t.Stop()
	var r reply
	select {
	case r = <-ch:
	case <-t.C:
		why := fmt.Sprintf("No acknowledgement within %s.", wait)
		if legacy {
			why = "This server predates ADR 0122 PR 5 and sends no acknowledgements, and no error came back."
		}
		return textResult("status: unknown",
			why+" The move may or may not have landed: call wait_for_decision to read the state."), nil
	case <-ctx.Done():
		return textResult("status: unknown", "Cancelled after sending. Call wait_for_decision to read the state."), nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	switch r.kind {
	case protocol.KindAck:
		if s.win == w {
			w.state = winActed
		}
		s.stats.decisionTimes = append(s.stats.decisionTimes, s.now().Sub(w.openedAt))
		return textResult("status: accepted", fmt.Sprintf("seq: %d", r.ack.Seq), "Call wait_for_decision for the next decision."), nil
	case protocol.KindError:
		s.stats.rejections++
		w.rejections++
		lines := []string{"status: rejected", "code: " + r.err.Code, "message: " + nameWrapperFor(s.view).apply(r.err.Message)}
		if r.err.Reason != "" {
			lines = append(lines, "reason: "+r.err.Reason)
		}
		if len(r.err.Missing) > 0 {
			lines = append(lines, "missing: "+strings.Join(r.err.Missing, ""))
		}
		if w.rejections >= maxRejections {
			lines = append(lines, "That is three refusals in this window: only an always-legal move is accepted now.")
		} else {
			lines = append(lines, "The window is still open; choose again from the same list.")
		}
		return textResult(lines...), nil
	default:
		return textResult("status: unknown", "The connection dropped before the server answered. Call wait_for_decision to read the state."), nil
	}
}

func nameWrapperFor(v *protocol.GameView) nameWrapper {
	if v == nil {
		return nameWrapper{}
	}
	return newNameWrapper(v)
}

// valueParamKey is the action parameter each open set fills: the keys
// legal writes for the same answers (choiceParams.CardName, and the cast /
// activate XValue), as legal.ValueCardName and legal.ValueX document.
var valueParamKey = map[string]string{
	legal.ValueCardName: "card_name",
	legal.ValueX:        "x_value",
}

// applyValue fills a move's open set (§6.2) from act's value, checked
// against the set the server stated. A move with no open set takes no
// value.
func applyValue(m legal.Move, value any) (json.RawMessage, error) {
	if value == nil {
		return m.Params, nil
	}
	if m.Value == nil {
		return nil, errors.New("this move takes no value; send it without one")
	}
	key := valueParamKey[m.Value.Kind]
	if key == "" {
		return nil, fmt.Errorf("this server marks the move with an open set (%q) this binary does not know", m.Value.Kind)
	}
	var filled any
	switch m.Value.Kind {
	case legal.ValueCardName:
		name, ok := value.(string)
		name = strings.TrimSpace(name)
		if !ok || name == "" || utf8.RuneCountInString(name) > maxCardNameValue {
			return nil, fmt.Errorf("value must be a card name of 1 to %d characters", maxCardNameValue)
		}
		filled = name
	case legal.ValueX:
		f, ok := value.(float64)
		if !ok || f != math.Trunc(f) {
			return nil, errors.New("value must be a whole number for X")
		}
		x := int(f)
		if (m.Value.Min != nil && x < *m.Value.Min) || (m.Value.Max != nil && x > *m.Value.Max) {
			return nil, fmt.Errorf("X=%d is outside the range the server stated", x)
		}
		filled = x
	}
	params := map[string]any{}
	if len(m.Params) > 0 {
		if err := json.Unmarshal(m.Params, &params); err != nil {
			return nil, err
		}
	}
	params[key] = filled
	return json.Marshal(params)
}

// --- say ----------------------------------------------------------------

// Say sends one chat line to the table.
func (s *Seat) Say(_ context.Context, in SayInput) (Result, error) {
	text := strings.TrimSpace(in.Text)
	if n := utf8.RuneCountInString(text); n == 0 || n > chatMaxRunes {
		return errorResult("text must be 1 to %d characters", chatMaxRunes), nil
	}
	s.mu.Lock()
	seated := s.sess != nil
	s.mu.Unlock()
	if !seated {
		return errorResult("not seated: call join first"), nil
	}
	if ok, wait := s.chat.allow(); !ok {
		return textResult("status: rate_limited", fmt.Sprintf("One chat line per 5 seconds: retry in %s.", wait.Round(100*time.Millisecond))), nil
	}
	s.mu.Lock()
	conn := s.conn
	s.mu.Unlock()
	if conn == nil {
		return errorResult("not connected"), nil
	}
	if err := conn.send(string(protocol.KindChat), "", protocol.ChatPayload{Text: text}); err != nil {
		return errorResult("not sent: %v", err), nil
	}
	return textResult("status: sent"), nil
}

// --- concede ------------------------------------------------------------

// Concede leaves the game (CR 104.3a). It is not a legal-move list entry,
// as it is not for the bot, so it has its own tool with a confirm flag.
func (s *Seat) Concede(ctx context.Context, in ConcedeInput) (Result, error) {
	if !in.Confirm {
		return errorResult("concede needs confirm: true"), nil
	}
	s.mu.Lock()
	sess, view := s.sess, s.view
	s.mu.Unlock()
	if sess == nil {
		return errorResult("not seated: call join first"), nil
	}
	if view == nil || view.State != "active" {
		return errorResult("there is no game in progress to concede"), nil
	}
	m := legal.Move{Type: "concede", Player: sess.PlayerID}
	ch, err := s.sendAction(m, nil)
	if err != nil {
		return errorResult("not sent: %v", err), nil
	}
	wait, _ := s.ackWait()
	t := time.NewTimer(wait)
	defer t.Stop()
	select {
	case r := <-ch:
		switch r.kind {
		case protocol.KindAck:
			return textResult("status: conceded"), nil
		case protocol.KindError:
			return errorResult("status: rejected\n%s", r.err.Message), nil
		}
		return textResult("status: unknown", "The connection dropped; call wait_for_decision to read the state."), nil
	case <-t.C:
		return textResult("status: unknown", "No acknowledgement yet; call wait_for_decision to read the state."), nil
	case <-ctx.Done():
		return textResult("status: unknown"), nil
	}
}

// --- leave --------------------------------------------------------------

// Leave disconnects and deletes the saved session. It refuses while the
// game is live and the seat is in it: a vanished seat holds the table.
func (s *Seat) Leave(_ context.Context, _ LeaveInput) (Result, error) {
	s.opMu.Lock()
	defer s.opMu.Unlock()
	s.mu.Lock()
	sess := s.sess
	if sess == nil {
		s.mu.Unlock()
		return errorResult("not seated"), nil
	}
	if s.haveView && s.view.State == "active" && s.endReason == "" {
		if self := s.selfLocked(); self != nil && !self.Eliminated {
			s.mu.Unlock()
			return errorResult("the game is live and your seat is in it: leaving would hold the table. Concede first (concede with confirm: true), then leave."), nil
		}
	}
	s.mu.Unlock()
	s.Close()
	if err := s.store.remove(sess.Origin, sess.GameID); err != nil {
		s.log.Warn("could not delete the saved session", "err", err)
	}
	s.mu.Lock()
	s.sess, s.view, s.haveView, s.win, s.moves = nil, nil, false, nil, nil
	s.key, s.lastCounted = windowKey{}, windowKey{}
	s.endReason, s.reportedOver, s.passUntil = "", false, false
	s.noMoveReq, s.probed = false, false
	s.autoSince, s.chatSince, s.logSeqShown = map[string]int{}, nil, 0
	s.broadcastLocked()
	s.mu.Unlock()
	return textResult("status: left", "Disconnected, and the saved session is deleted."), nil
}
