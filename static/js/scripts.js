var allMessages = {};
var lastRenderedIds = {
  chatgpt: null,
  rookie: null,
  interviewer: null,
};

var socket = new WebSocket("ws://localhost:8080/ws");

function getInitials(name) {
  return name
    .split(" ")
    .map((word) => word[0].toUpperCase())
    .join("");
}

marked.setOptions({ breaks: true, gfm: true });

socket.onmessage = function (event) {
  console.log("Received message:", event.data);
  try {
    var data = JSON.parse(event.data);
    updateOrCreateMessage(data);
    renderMessage(data);
  } catch (error) {
    console.error("Error parsing message:", error);
  }
};

function updateOrCreateMessage(data) {
  if (!allMessages[data.list_name]) {
    allMessages[data.list_name] = {};
  }

  if (!allMessages[data.list_name][data.id]) {
    allMessages[data.list_name][data.id] = data;
  } else {
    allMessages[data.list_name][data.id].text = data.text;
  }
}

function renderMessage(data) {
  const chatContent = document.getElementById("chatContent");
  let messageElement = document.querySelector(
    `.message[data-id="${data.id}"][data-list-name="${data.list_name}"]`
  );

  if (!messageElement) {
    messageElement = createMessageElement(data);
    chatContent.appendChild(messageElement);
  }

  const textElement = messageElement.querySelector(".message-text");
  updateMessageContent(textElement, data.text);

  lastRenderedIds[data.list_name.toLowerCase()] = data.id;

  chatContent.scrollTop = chatContent.scrollHeight;
}

function updateMessageContent(element, newText) {
  if (!element) {
    console.error("Element is null in updateMessageContent");
    return;
  }

  const currentText = element.getAttribute("data-full-text") || "";
  if (newText !== currentText) {
    element.setAttribute("data-full-text", newText);
    const additionalText = newText.slice(currentText.length);
    if (additionalText) {
      typeWriterEffect(element, additionalText);
    }
  }
}

function typeWriterEffect(element, text) {
  let index = 0;
  element.classList.add("typing");

  function type() {
    if (index < text.length) {
      element.innerHTML += text.charAt(index);
      index++;
      setTimeout(type, 5);
    } else {
      element.classList.remove("typing");
      element.innerHTML = marked.parse(element.getAttribute("data-full-text"));
    }
  }

  type();
}

function createMessageElement(data) {
  var messageElement = document.createElement("div");
  messageElement.className = "message " + data.list_name.toLowerCase();
  messageElement.setAttribute("data-id", data.id);
  messageElement.setAttribute("data-list-name", data.list_name);

  var avatarElement = document.createElement("div");
  avatarElement.className = "avatar";
  avatarElement.textContent = getInitials(data.list_name);

  var contentElement = document.createElement("div");
  contentElement.className = "message-content";

  var roleElement = document.createElement("div");
  roleElement.className = "role";
  roleElement.textContent = data.list_name;

  var idElement = document.createElement("span");
  idElement.className = "id";
  idElement.textContent = " (ID: " + data.id + ")";
  roleElement.appendChild(idElement);

  var textElement = document.createElement("div");
  textElement.className = "message-text markdown-body";

  contentElement.appendChild(roleElement);
  contentElement.appendChild(textElement);
  messageElement.appendChild(avatarElement);
  messageElement.appendChild(contentElement);

  return messageElement;
}

socket.onopen = function (event) {
  console.log("WebSocket connection opened");
};

socket.onclose = function (event) {
  console.log("WebSocket connection closed");
};

socket.onerror = function (error) {
  console.error("WebSocket error:", error);
};
