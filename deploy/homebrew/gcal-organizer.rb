class GcalOrganizer < Formula
  desc "Automate meeting note organization, calendar syncing, and task assignment"
  homepage "https://github.com/jflowers/gcal-organizer"
  url "https://github.com/jflowers/gcal-organizer/archive/refs/tags/v1.1.0.tar.gz"
  sha256 "PLACEHOLDER_SHA256"
  license "MIT"
  head "https://github.com/jflowers/gcal-organizer.git", branch: "main"

  depends_on "go" => :build
  depends_on "node" # Required for browser-based task assignment (Playwright)

  def install
    # Build the Go binary
    system "go", "build", *std_go_args(ldflags: "-s -w"), "./cmd/gcal-organizer"

    # Install man page
    man1.install "man/gcal-organizer.1"

    # Install browser automation scripts
    (libexec/"browser").install Dir["browser/*"]

    # Install service wrapper
    (libexec/"deploy").install "deploy/run-wrapper.sh"
    chmod 0755, libexec/"deploy/run-wrapper.sh"

    # Install browser dependencies
    cd libexec/"browser" do
      system "npm", "install", *std_npm_args(prefix: false)
    end
  end

  def post_install
    # Create config directory
    (var/"gcal-organizer").mkpath
  end

  service do
    run [opt_libexec/"deploy/run-wrapper.sh"]
    run_type :interval
    interval 3600
    log_path var/"log/gcal-organizer.log"
    error_log_path var/"log/gcal-organizer.log"
    environment_variables PATH: std_service_path_env,
                          GCAL_DAYS_TO_LOOK_BACK: "1",
                          GCAL_ORGANIZER_BIN: "#{HOMEBREW_PREFIX}/bin/gcal-organizer",
                          HOME: Dir.home
  end

  def caveats
    <<~EOS
      To get started:

        1. Set up Google Cloud credentials:
           See: https://github.com/jflowers/gcal-organizer/blob/main/docs/SETUP.md

        2. Create your env file:
           mkdir -p ~/.gcal-organizer
           cp #{etc}/gcal-organizer/.env.example ~/.gcal-organizer/.env
           # Edit ~/.gcal-organizer/.env with your API keys

        3. Authenticate:
           gcal-organizer auth login

        4. Test with dry-run:
           gcal-organizer run --dry-run --verbose

        5. Start the hourly service:
           brew services start gcal-organizer

      Man page: man gcal-organizer
    EOS
  end

  test do
    assert_match "gcal-organizer", shell_output("#{bin}/gcal-organizer --help")
  end
end
