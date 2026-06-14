# logger

WIP. slog driver for this module. Notes to format later:

- Implements the "logger in the context" pattern: the logger travels inside
  context.Context. Store it once at startup with WithLogger(ctx, client); retrieve
  it anywhere with FromContext(ctx). FromContext never returns nil (falls back to a
  wrapped slog.Default()).
- Client interface exposes only the ...Context methods (InfoContext, ErrorContext,
  DebugContext, WarnContext) because they forward ctx to the handler (needed for
  future OpenTelemetry correlation).
- NewSlogLogger returns the concrete *SlogLogger (not Client) so the caller keeps
  access to SetLevel / SetDefaultLevel. Level changes go through a shared
  slog.LevelVar and apply to every record with no rebuild.
- The config must already be generated and validated. NewSlogLogger does NOT
  validate its input; it assumes a *slogconfig.Config produced by
  slogconfig.NewConfig() (github.com/a-castellano/go-types/slog). Passing an invalid
  or nil config is a programming error and is not guarded against.
- Output is always stdout. systemd / Podman-under-systemd + journald manage the
  final destination. Severity lives in the structured level field.
