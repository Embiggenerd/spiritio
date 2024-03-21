/**
 * @type {import("../../types").MediaService}
 */
const media = {
    permissionsGranted: false,
    peerConnection: null,
    constraints: {
        video: true,
        audio: true,
    },
    stream: null,
    async init() {
        try {
            // Ask permission to access mic and cam devices
            this.stream = await navigator.mediaDevices.getUserMedia(
                this.constraints
            )
            this.permissionsGranted = true
            this.peerConnection = new RTCPeerConnection()
            return this
        } catch (e) {
            // If rejected, return with permissionsGranted = false
            return this
        }
    },
    closePeerConnection: function () {
        if (this.stream)
            this.stream.getTracks().forEach((s) => {
                s.stop()
            })
    },
    createAnswer: function () {
        if (this.peerConnection) return this.peerConnection.createAnswer()
    },
    addCandidate: function (candidate) {
        if (this.peerConnection) this.peerConnection.addIceCandidate(candidate)
    },
    setRemoteDescription: function (offer) {
        if (this.peerConnection) this.peerConnection.setRemoteDescription(offer)
    },
    addTrack: function () {
        if (this.stream) {
            this.stream.getTracks().forEach((track) => {
                if (this.peerConnection && this.stream) {
                    this.peerConnection.addTrack(track, this.stream)
                }
            })
        }
    },
    assignCallbacks: function (trackHandler, iceCandidateHandler) {
        if (this.peerConnection) {
            this.peerConnection.ontrack = trackHandler
            this.peerConnection.onicecandidate = iceCandidateHandler
        }
    },
    setLocalDescription: function (answer) {
        if (this.peerConnection) this.peerConnection.setLocalDescription(answer)
    },
}

export default media
