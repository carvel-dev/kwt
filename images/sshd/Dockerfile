FROM       ubuntu:xenial
MAINTAINER Dmitriy Kalinin "https://github.com/cppforlife"

RUN apt-get update && apt-get install -y openssh-server
RUN mkdir /var/run/sshd

# Add user 'tom' for running and logging in
RUN useradd -ms /bin/bash tom
RUN mkdir /home/tom/.ssh && chmod 700 /home/tom/.ssh
RUN touch /home/tom/.ssh/authorized_keys && chmod 644 /home/tom/.ssh/authorized_keys

# Configure sshd_config
RUN sed "/^ *Port/d" -i /etc/ssh/sshd_config && echo 'Port 2048' >> /etc/ssh/sshd_config
ADD harden.sh /tmp/harden.sh
RUN chmod +x /tmp/harden.sh && /tmp/harden.sh

RUN chown tom -R /home/tom /etc/ssh
RUN apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

EXPOSE 2048
USER tom
CMD ["/usr/sbin/sshd", "-D"]
