

class ProfileComponent extends HTMLElement {
  constructor() {
    super();

    // Attach a shadow DOM to the element.
    const shadow = this.attachShadow({ mode: 'open' });


    // Apply some styles to the wrapper

    const template = document.createElement('template')

    template.innerHTML = `
      <style>

      #header-container {
        display: flex;
        justify-content: space-around;
        gap: 60px;
        align-items: center;
      }

      ::slotted([slot="header"]) {
        position: fixed;
        top: 15px;
        right: 10px;
        font-family: Aruma;
        font-size: 40px;
        border: 3px solid black;
        background-color: lightsalmon;
        border-radius: 10px;
        padding: 25px;
        text-align: center;
        box-shadow: 5px 5px black;
      }


      ::slotted([slot="body-1"]) {
        flex-grow: 3;
        border: 3px solid black;
        background-color: seagreen;
        font-size: 40px;
        font-family: Cheto;
        padding: 20px;
        border-radius: 10px;
        box-shadow: 5px 5px black;
      }
      
      .break-line {
        height: 30px;
        margin-top: 30px;
        margin-bottom: 30px;
        display: flex;
        justify-content: space-between;
        align-items: center;
      }

      #break-line-left, #break-line-right {
        background-color: black;
        border-radius:5px;
        width: 45%;
        height: 10px;
      }


      .five-star{
        width: 0;
        height:0;
        border-left: solid 25px transparent;
        border-right: solid 25px transparent;
        border-bottom: solid 17.5px red;
        transform: rotatez(35deg);
      }



      .five-star:before{
        position: absolute;
        display:block;
        width: 0;
        height:0;
        top: -11.75px;
        left: -16.75px;
        border-left: solid 7.5px transparent;
        border-right: solid 7.5px transparent;
        border-bottom: solid 20px red;
        transform: rotatez(-35deg);
        content:'';
      }

      .five-star:after{
        position:absolute;
        display:block;
        width: 0;
        height:0;
        top:0.75px;
        left:-26.75px;
        border-left: solid 25px transparent;
        border-right: solid 25px transparent;
        border-bottom: solid 17.5px red;
        transform: rotatez(-70deg);
        content:'';
      }
      </style>
      <div>
        <div id="header-container">
          <!-- <slot name="header">ADD HEADER</slot> -->
          <slot name="body-1">ADD BODY-1</slot>
        </div>
          <div class="break-line">
          <div id="break-line-left"></div>
          <div id="five-star-container">
          <div class="five-star"></div>
          </div>
          <div id="break-line-right"></div>
        </div>
      </div>
    `;




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


customElements.define('profile-component', ProfileComponent);
