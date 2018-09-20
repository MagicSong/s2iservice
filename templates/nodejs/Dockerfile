FROM fortinj66/centos7-s2i-nodejs:8.9.4

ENV NPM_REGISTRY="http://nexus.kubesphere.com/repository/npm_proxy/"
RUN npm config set registry ${NPM_REGISTRY}