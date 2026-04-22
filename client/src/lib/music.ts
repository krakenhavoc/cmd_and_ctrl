// Ambient music playback — a single long-form track that loops
// forever as the app shell's background bed.
//
// Deliberately a separate module from sounds.ts because the SFX
// model (clone-on-play pools, per-event rate limits, random variant
// picker) doesn't translate: music is one element, always the same
// URL, and only ever has one playing instance. Sharing the pool
// machinery would be more code, not less.
//
// Autoplay policy: same constraint as SFX — HTMLAudioElement.play()
// rejects until the page sees a user gesture. armMusicOnFirstGesture()
// installs a one-shot pointerdown/keydown listener that starts the
// track. Safe to call many times.
//
// Volume/mute: routed through settings.ts the same way SFX volume
// is. master × music (both 0..100 in the store) folds into a 0..1
// multiplier here; the audio.muted toggle pauses/resumes the track
// (vs. the SFX path which just no-ops play()). We pause rather than
// zero the volume so a muted user isn't silently decoding audio
// they'll never hear.
//
// Dependency-free + SSR-safe (window-guarded) so it can be imported
// from any surface.

const AMBIENT_TRACK_URL = "/sounds/ambient_bronze_thunder.mp3";

let audio: HTMLAudioElement | null = null;
let armed = false;
// musicMultiplier is master × music volume, both 0..100 in the
// settings store, normalised to 0..1 here. The settings bridge
// (settings.ts) pushes updates via setMusicVolumeMultiplier.
let musicMultiplier = 0;
let muted = false;

function hasWindow(): boolean {
  return typeof window !== "undefined";
}

function applyVolume(): void {
  if (!audio) return;
  // Pause when muted so we're not decoding a silent stream in the
  // background tab. Unpause picks back up at the last position,
  // which is the right UX for an ambient bed.
  audio.volume = musicMultiplier;
  if (muted || musicMultiplier === 0) {
    if (!audio.paused) audio.pause();
  } else if (armed && audio.paused) {
    void audio.play().catch(() => {});
  }
}

export function setMusicVolumeMultiplier(v: number): void {
  musicMultiplier = Math.min(1, Math.max(0, v));
  applyVolume();
}

export function setMusicMuted(m: boolean): void {
  muted = m;
  applyVolume();
}

// armMusicOnFirstGesture attaches a one-shot pointerdown + keydown
// listener to window. On first of either, it creates the Audio
// element and starts playback. Subsequent calls no-op so the App
// shell can re-run it from a reactive effect without side effects.
export function armMusicOnFirstGesture(): void {
  if (armed || !hasWindow()) return;
  armed = true;
  const start = () => {
    if (!audio) {
      audio = new Audio(AMBIENT_TRACK_URL);
      audio.loop = true;
      audio.preload = "auto";
    }
    applyVolume();
    if (!muted && musicMultiplier > 0) {
      void audio.play().catch(() => {});
    }
    window.removeEventListener("pointerdown", start);
    window.removeEventListener("keydown", start);
  };
  window.addEventListener("pointerdown", start, { once: true });
  window.addEventListener("keydown", start, { once: true });
}
