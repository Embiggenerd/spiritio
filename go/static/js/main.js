import render from './renderer.js'
import media from './entities/media.js'
import message from './entities/message.js'
import appComponent from './component.js'

try {
    await appComponent.init(render, message, media)
} catch (e) {
    console.log(e)
    renderer.chatLog.addMessage({
        text: e,
        from: 'ADMIN (to you)',
    })
}
