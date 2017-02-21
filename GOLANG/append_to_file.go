out, err := os.OpenFile("/tmp/dh-client-env.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
if err != nil {
    if !os.IsExist(err) {
        out, err = os.Create("/tmp/dh-client-env.log")
        if err != nil {
            log.Fatal(err.Error())
        }
    } else {
        log.Fatal(err.Error())
    }
}

out.WriteString("\n--------------------- " + time.Now().Format("Mon, 2 Jan 2006 15:04:05 MST") + " ---------------------\n")

err = out.Close()
if err != nil {
    log.Fatal(err.Error())
}