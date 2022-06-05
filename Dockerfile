FROM busybox

COPY build/nfs-csi /

ENTRYPOINT ["/nfs-csi"]