FROM tinygo/tinygo:0.38.0

USER root
WORKDIR /root
ENTRYPOINT [ "/bin/bash" ]

RUN apt-get update
RUN DEBIAN_FRONTEND=noninteractive apt-get install -y wget curl pip

RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
RUN . /root/.bashrc && rustup target add wasm32-wasip1

RUN curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.3/install.sh | bash
RUN . /root/.bashrc && nvm install 22

ADD image-repo/requirements.txt /tmp
RUN pip install --break-system-packages -r /tmp/requirements.txt && rm /tmp/requirements.txt

RUN mkdir -p /root/spin && \
    cd /root/spin && \
    curl -fsSL https://spinframework.dev/downloads/install.sh | bash
RUN echo 'PATH=/root/spin:$PATH' >> /root/.bashrc

RUN . /root/.bashrc && \
    spin plugins install -y --url https://github.com/fermyon/spin-trigger-cron/releases/download/canary/trigger-cron.json && \
    spin templates install --git https://github.com/fermyon/spin-trigger-cron

# Warmup build
RUN . /root/.bashrc && \
    cd /tmp && \
    git clone https://github.com/mayflower/spin-workshop-2025 && \
    cd spin-workshop-2025/image-repo && \
    spin build && \
    cd /tmp && \
    rm -fr /tmp/spin-workshop-2025

