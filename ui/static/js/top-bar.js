

class TopBar extends HTMLElement {
  constructor() {
    super();

    // Attach a shadow DOM to the element.
    const shadow = this.attachShadow({ mode: 'open' });

    const template = document.createElement("template")

    template.innerHTML = `
    <div id="top-container">
        <div id="logo"> 
            <div id="logo-text">Ask Away</div>    
        </div>
        <div id="top-blue-bar"></div>
        <div id="top-profile-name">John</div>
    </div>
`


    const style = document.createElement("style")
    style.textContent = `
      #top-profile-name {
        margin-left: 20px;
        margin-right: 15px;
        font-family: Aruma;
        font-size: 40px;
        border: 3px solid black;
        background-color: lightsalmon;
        border-radius: 10px;
        padding: 15px;
        text-align: center;
        box-shadow: 5px 5px black;
      }

      #top-container {
        position: absolute;
        left: -5px;
        width:100%;
        height: 90px;
        display: flex;
        align-items: center;
      }
        #logo {
          position: absolute;
          left: 15px;
          background-color: peachpuff;
          width: 240px;
          height: 90px;
          border: 5px solid black;
          font-family: Parson;
          font-style: italic;
          font-size: 50px;
          text-align: center;
          box-shadow: 5px 5px black;
          border-radius: 10px;
          align-items: center;
          display: flex;
          justify-content: center;
          transform: skew(-5deg, -5deg);
        }
        #top-blue-bar {
          border-radius: 0 10px 10px 0px;
          background-color: skyblue;
          height: 65%;
          flex-grow: 1;
          border-top:  3px solid black;
          border-right:  4px solid black;
          border-bottom:  4px solid black;
          box-shadow: 5px 5px black;
        }
        #logo-text {
          text-align: center;
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


customElements.define('top-bar', TopBar);



