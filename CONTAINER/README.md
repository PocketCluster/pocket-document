# Container/Image Handling Tips

### (02/07/2017)

**How to configure locales to Unicode in a Docker container?**

The `/etc/default/locale` file is loaded by PAM; see `/etc/pam.d/login` for example. However, PAM is not invoked when running a command in a Docker container. To configure the locale, simply set the relevant environment variable in your Dockerfile. 

```sh
# Set the locale (We don't need to run locale-gen all the time)
RUN locale-gen en_US.UTF-8  
ENV LANG en_US.UTF-8  
ENV LANGUAGE en_US:en  
ENV LC_ALL en_US.UTF-8
```
Or simply one-liner

```sh
# Set the locale (We don't need to run locale-gen all the time)
RUN locale-gen en_US.UTF-8
ENV LANG='en_US.UTF-8' LANGUAGE='en_US:en' LC_ALL='en_US.UTF-8'
```

> Reference

- <http://askubuntu.com/questions/581458/how-to-configure-locales-to-unicode-in-a-docker-ubuntu-14-04-container>
- <http://serverfault.com/questions/362903/how-do-you-set-a-locale-non-interactively-on-debian-ubuntu/689947>
- <http://stackoverflow.com/questions/28405902/how-to-set-the-locale-inside-a-docker-container>
- <http://unix.stackexchange.com/questions/246846/cant-generate-en-us-utf-8-locale>