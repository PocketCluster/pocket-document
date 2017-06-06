<<<<<<< HEAD
        printErrors := func() {
            n, _ := io.Copy(os.Stderr, proxyErr)
            if n > 0 {
                os.Stderr.WriteString("\n")
            }
        }
=======
>>>>>>> 5cfd89fd... Merge pull request #728 from gravitational/ev/login-ux
