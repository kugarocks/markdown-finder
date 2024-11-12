class MarkdownFinder < Formula
    desc "Markdown Finder in your terminal"
    homepage "https://github.com/kugarocks/markdown-finder"
  
    if Hardware::CPU.arm?
      url "https://github.com/kugarocks/markdown-finder/releases/download/v1.0.0/mdf_1.0.0_darwin_arm64.zip"
      sha256 "79428bf2fd27bf9be9df3bee27ec143ae30992208afd61d1eae514a467080cf0"
    else
      url "https://github.com/kugarocks/markdown-finder/releases/download/v1.0.0/mdf_1.0.0_darwin_amd64.zip"
      sha256 "61fd41e66e2df63d243b30293f0ca95b18236edc3e44822f96ff74263c3386e5"
    end
  
    def install
      bin.install "mdf"
    end
  
    test do
      system "#{bin}/mdf", "--version"
    end
  end
