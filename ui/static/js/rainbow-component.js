
class RainbowComponent extends HTMLElement {
  constructor() {
    super();

    // Attach a shadow DOM to the element.
    const shadow = this.attachShadow({ mode: 'open' });

    const template = document.createElement("template")

    template.innerHTML = `
      <div id="box-container">
        <div class="box">
        </div>
        <div class="box">
        </div>
        <div class="box">
        </div>
        <div class="box">
        </div>
        <div class="box">
        </div>
        <div class="box">
        </div>
        <div class="box">
        </div>
      </div>
`


    const style = document.createElement("style")
    style.textContent = `

@keyframes fall-down{
  from {
    top: 0;
    z-index: 0;
  }

  to {
    top: 100vh;
    z-index: 10;
  }
}

#box-container {
  position: fixed;
  // top: -20px;
}
  

.box:nth-of-type(1) {
  --box-index: 1;
  background-color: yellow;
  // animation-delay: 0.5s;
}

.box:nth-of-type(2) {
  --box-index: 2;
  background-color: orange ;
  animation-delay: 0.25s;
}

.box:nth-of-type(3) {
  --box-index: 3;
  background-color: red;
  animation-delay: 0.5s
}

.box:nth-of-type(4) {
  --box-index: 4;
  background-color: purple;
  animation-delay: .75s;
}

.box:nth-of-type(5) {
  --box-index: 5;
  background-color: blue;
  animation-delay: 1s;
}

.box:nth-of-type(6) {
  --box-index: 6;
  background-color: green;
  animation-delay: 1.25s;
}
.box:nth-of-type(7) {
  --box-index: 7;
  background-color: maroon;
  animation-delay: 1.5s;
}
.box:nth-of-type(8) {
  --box-index: 8;
  background-color: fuchsia;
  animation-delay: 1.75s;
}


.box {
  animation: 3s ease-in fall-down infinite;
  width: 100vw;
  height: 100vh;
  position: absolute;
}
`

    template.style = style

    shadow.appendChild(style)
    shadow.appendChild(template.content.cloneNode(true))

  }
  connectedCallback() {


    console.log("Custom element added to page.");
  }

  disconnectedCallback() {
    console.log("Custom element removed from page.");
  }

  adoptedCallback() {
    console.log("Custom element moved to new page.");
  }

  attributeChangedCallback(name, oldValue, newValue) {
    console.log(`Attribute ${name} has changed.`);
  }
}


customElements.define('rainbow-component', RainbowComponent);



