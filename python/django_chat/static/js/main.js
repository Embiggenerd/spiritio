let chatName = 'message'
let chatSocket = null
let chatWindowURL = window.location.href
let chatRoomUuid = Math.random().toString(36).slice(2, 12)


const chatElement = document.querySelector('#chat')
const chatOpenElement = document.querySelector('#chat_open')
const chatJoinElement = document.querySelector('#chat_join')
const chatIconElement = document.querySelector('#chat_icon')
const chatWelcomeElement = document.querySelector('#chat_welcome')
const chatRoomElement = document.querySelector('#chat_room')
const chatNameElement = document.querySelector('#chat_name')
const chatLogElement = document.querySelector('#chat_log')
const chatInputElement = document.querySelector('#chat_message_input')
const chatSubmitElement = document.querySelector('#chat_message_submit')

function getCookie(name) {
    let cookieValue = null
    if(document.cookie && document.cookie !== '') {
        const cookies = document.cookie.split(';')
        let i = 0
        while(i < cookies.length){
            const cookie = cookies[i].trim()
            if(cookie.substring(0, name.length + 1) === (name + '=')) {
                cookieValue = decodeUTIComponent(cookie.substring(name.length + 1))
                break
            }
            i++
        }
    }
    console.log('cookieValue', cookieValue)
    return cookieValue
}

function scrollToBottom() {
    chatLogElement.scrollTop = chatLogElement.scrollHeight
}

function sendMessage() {
    chatSocket.send(JSON.stringify({
        'type': 'message',
        'message': chatInputElement.value,
        'name': chatName,
    }))
    chatInputElement.value = ''
}



function onChatMessage(data) {
    console.log(
        'onMessage', data
    )
    if (data.type == 'chat_message') {
        if (data.agent) {
            chatLogElement.innerHTML += `
                <div class="flex w-full mt-2 space-x-3 max-w-md">
                    <div class="flex-shrink-0 h-10 w-10 rounded-full bg-gray-300 text-center pt-2">${data.initials}</div>

                    <div>
                        <div class="bg-gray-300 p-3 rounded-l-lg rounded-br-lg">
                            <p class="text-sm">${data.message}</p>
                        </div>
                        
                        <span class="text-xs text-gray-500 leading-none">${data.created_at} ago</span>
                    </div>
                </div>
            `

        } else {
            chatLogElement.innerHTML += `
            <div class="flex w-full mt-2 space-x-3 max-w-md ml-auto justify-end">
                <div>
                    <div class="bg-blue-300 p-3 rounded-l-lg rounded-br-lg">
                        <p class="text-sm">${data.message}</p>
                    </div>
                    
                    <span class="text-xs text-gray-500 leading-none">${data.created_at} ago</span>
                </div>

                <div class="flex-shrink-0 h-10 w-10 rounded-full bg-gray-300 text-center pt-2">${data.initials}</div>
            </div>
        `    
        }
    } else if (data.type == 'users_update') {
        chatLogElement.innerHTML += '<p class="mt-2"> An admin has joined the room</p>'
    }
    scrollToBottom()
}

async function joinChatRoom(){
    console.log('joined chat room')
    chatName = chatNameElement.value
    console.log('hihi', chatName)
    console.log('room uuid', chatRoomUuid)
    const data = new FormData()
    data.append('name', chatName)
    data.append('url', chatWindowURL)
    await fetch(`/api/create-room/${chatRoomUuid}/`, {
        method: 'POST',
        headers: {
            // 'X-CSRFToken': getCookie('csrftoken')
        },
        body: data,
    })
    .then(function(res) {
        
        
        return res.json()
    })
    .then(function(data) {
        console.log('data', data)
    })
    .catch(function(e){
        console.log(e)
    })

    chatSocket = new WebSocket(`ws://${window.location.host}/ws/${chatRoomUuid}/`)

    chatSocket.onmessage = function(e) {
        console.log('onMessage')
        onChatMessage(JSON.parse(e.data))
    }
    chatSocket.onopen = function(e) {
        scrollToBottom()
    }
    chatSocket.onclose = function(e) {
        console.log('onClose')
    }
}


chatOpenElement.onclick = function(e) {
    e.preventDefault()
    chatIconElement.classList.add('hidden')
    chatWelcomeElement.classList.remove('hidden')
    return false
}

chatJoinElement.onclick = function(e) {
    e.preventDefault()
    chatWelcomeElement.classList.add('hidden')
    chatRoomElement.classList.remove('hidden')
    joinChatRoom()
    
    
}

chatSubmitElement.onclick = function(e) {
    e.preventDefault()
    sendMessage()
    return false
}

