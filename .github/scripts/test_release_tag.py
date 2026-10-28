"""Exercise the actual workflow tag selector against real Git tags."""

import os
from pathlib import Path
import subprocess
import tempfile
import unittest


class ReleaseTagTest(unittest.TestCase):
    def test_latest_stable_tag(self):
        workflow = Path(__file__).resolve().parents[1] / "workflows/fork-fallback.yml"
        selector = next(line.strip() for line in workflow.read_text().splitlines()
                        if '[ -n "$tag" ] || tag=' in line)
        cases = [
            (["1.9.1", "1.10.1"], "1.10.1"),
            (["1.10.1", "v1.10.2"], "v1.10.2"),
            (["v1.9.1", "1.10.2"], "1.10.2"),
            (["v1.9.1", "v1.10.2", "1.10.3-rc.1", "1.10.9-fallback.1",
              "release-99.0.0", "v1.10.99junk"], "v1.10.2"),
        ]
        for tags, expected in cases:
            with self.subTest(tags=tags), tempfile.TemporaryDirectory() as directory:
                env = dict(os.environ, GIT_AUTHOR_NAME="test", GIT_AUTHOR_EMAIL="test@example.com",
                           GIT_COMMITTER_NAME="test", GIT_COMMITTER_EMAIL="test@example.com")
                def git(*args):
                    subprocess.run(["git", *args], cwd=directory, env=env, check=True,
                                   stdout=subprocess.DEVNULL, stderr=subprocess.PIPE)
                git("init")
                git("commit", "--allow-empty", "-m", "fixture")
                for tag in tags:
                    git("tag", tag)
                result = subprocess.run(
                    ["bash", "-eo", "pipefail", "-c", 'tag=""\n' + selector + '\nprintf "%s" "$tag"'],
                    cwd=directory, env=env, check=True, capture_output=True, text=True)
                self.assertEqual(result.stdout, expected)


if __name__ == "__main__":
    unittest.main()
