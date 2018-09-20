FROM appuio/s2i-maven-java

ADD settings.xml /opt/app-root/src/.m2/settings.xml
USER root
ENV NEXUS_ADDRESS=nexus.kubesphere.com
RUN chown -R 1001:1001 /opt/app-root/src/.m2
USER 1001
CMD ["/usr/libexec/s2i/usage"]