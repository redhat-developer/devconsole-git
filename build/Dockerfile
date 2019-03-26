FROM alpine:3.8

<<<<<<< HEAD
RUN apk upgrade --update --no-cache

USER nobody

ADD build/_output/bin/git-operator /usr/local/bin/git-operator
=======
ENV OPERATOR=/usr/local/bin/git-service2 \
    USER_UID=1001 \
    USER_NAME=git-service2

# install operator binary
COPY build/_output/bin/git-service2 ${OPERATOR}

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
>>>>>>> baseline for git push CR
