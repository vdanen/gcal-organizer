# Feature Specification: Hourly Service Deployment

**Feature Branch**: `004-service-deployment`  
**Created**: 2026-02-07  
**Status**: Draft  
**Input**: Run gcal-organizer hourly as a service on Fedora and macOS with 1-day lookback

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Hourly Scheduling on macOS (Priority: P1)

As a macOS user, I want gcal-organizer to run automatically every hour so my meeting documents, calendar attachments, and task assignments stay organized without manual intervention.

**Why this priority**: Primary development platform

**Independent Test**: Install the service, wait 1 hour, check logs for a successful run

**Acceptance Scenarios**:

1. **Given** the service is installed via `make install-service`, **When** 1 hour elapses, **Then** `gcal-organizer run` executes with 1-day lookback and logs output
2. **Given** the service is running, **When** I run `make service-status`, **Then** I see the current state and last run time
3. **Given** I want to stop it, **When** I run `make uninstall-service`, **Then** the service is removed cleanly

---

### User Story 2 - Hourly Scheduling on Fedora (Priority: P1)

As a Fedora user, I want the same hourly scheduling via systemd user services.

**Why this priority**: Secondary platform, equal importance

**Independent Test**: Install the systemd timer, wait 1 hour, check `journalctl --user` for output

**Acceptance Scenarios**:

1. **Given** the service is installed via `make install-service`, **When** 1 hour elapses, **Then** `gcal-organizer run` executes with 1-day lookback and logs output
2. **Given** the service is running, **When** I run `systemctl --user status gcal-organizer.timer`, **Then** I see the timer state and next trigger time
3. **Given** I want to stop it, **When** I run `make uninstall-service`, **Then** the systemd units are disabled and removed

---

### User Story 3 - Log Output (Priority: P2)

As a user, I want to review logs from previous runs to troubleshoot issues.

**Why this priority**: Debugging aid

**Acceptance Scenarios**:

1. **Given** the service has run, **When** I check logs, **Then** output includes timestamps and full workflow summary
2. **On macOS**: Logs available via `log show --predicate 'subsystem=="com.jflowers.gcal-organizer"'` or log file
3. **On Fedora**: Logs available via `journalctl --user -u gcal-organizer.service`

---

### Edge Cases

- What happens when the browser automation (task assignment) runs without a display?
  → On macOS, launchd runs in the user session and has access to the GUI. On Fedora, the systemd user service runs in the user session; if no display is available, Step 3 (assign-tasks) is skipped gracefully
- What happens when the machine is asleep during a scheduled run?
  → macOS launchd will run the job when the machine wakes. systemd timers with `Persistent=true` catch up on missed runs
- What happens when credentials expire?
  → The service logs the auth error and exits; user re-authenticates manually
- What happens on network failure?
  → API errors are logged, service exits non-zero, runs again next hour

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Service MUST run `gcal-organizer run` with `GCAL_DAYS_TO_LOOK_BACK=1` every hour
- **FR-002**: Service MUST use the user's existing OAuth token and credentials
- **FR-003**: macOS service MUST be a launchd user agent (LaunchAgent)
- **FR-004**: Fedora service MUST be a systemd user service with timer
- **FR-005**: Service MUST log output with timestamps for troubleshooting
- **FR-006**: Service MUST provide install/uninstall commands via Makefile targets
- **FR-007**: Service MUST catch up on missed runs (macOS wake, systemd Persistent=true)

### Configuration Requirements

- **CR-001**: Environment variables loaded from `~/.gcal-organizer/.env` or system environment
- **CR-002**: Binary path auto-detected or configurable

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Service runs successfully every hour without manual intervention
- **SC-002**: `make install-service` completes in under 5 seconds
- **SC-003**: Logs from the last 24 hours are easily retrievable
- **SC-004**: Service survives system reboot (auto-starts on login)
