var Gotcha = {
  id: null,
  clientId: null,
  audioURL: null,
  imageURL: null,
  expiry: null,
  init: function(clientId, opts) {
    // defaults
    var language = "en";
    var nogzip = false;
    if (opts !== null && opts.language !== null) {
      language = opts.language;
    }
    if (opts !== null && opts.nogzip !== null) {
      nogzip = opts.nogzip;
    }
    return fetch("{{.PublicURL}}/gotcha/register", {
      method:"POST", 
      body: {
        "client-id": clientId
      }
    }).then(function(response) {
      return response.json()
    }).then(function(data) {
      this.id = data.id;
      this.clientId = clientId;
      this.audioURL = data["audio-url"];
      this.imageURL = data["image-url"];
      this.expiry = new Date(data.expiry);
    }).then(function() {
      return this;
    })
  },
  refresh: function() {
    fetch(`{{.PublicURL}}/gotcha/refresh`, {
      method: 'POST',
      body: {
        id: this.id
      }
    }).then(function(response) {
      return response.json()
    }).then(function(data) {
      this.audioURL = data["audio-url"];
      this.imageURL = data["image-url"];
    })
  },
  render: function(id, opts) {
    var div = createElement('div', {
      class: "gotcha-captcha",
      id: "gotcha-challenge-"+this.id,
      "data-gotcha-id": this.id
    });
    div.appendChild(createElement('img', {
      "id": "gotcha-challenge-image",
      "src": "{{.PublicURL}}/gotcha/assets/"+this.id+"/image.png",
      "alt": "gotcha captcha challenge image",
    }));
    div.appendChild(createElement('audio', {
      "id": "gotcha-challenge-audio",
      "src": "{{.PublicURL}}/gotcha/assets/"+this.id+"/audio.wav",
      "type": "audio/wav"
    }));
    div.appendChild(createElement('input', {
      "id": "gotcha-challenge-response",
      "name":"gotcha-challenge-response"
    }));
    var btnGroup = createElement('div', {
      "class": "gotcha-button-group",
      "style": "width: 100%;"
    });
    var valButton = createElement('button', {
      "class": "gotcha-button gotcha-validate",
      "type": "button",
    });
    valButton.innerHTML = "Validate";
    btnGroup.appendChild(valButton);
    var refresh = createElement('button', {
        "class": "gotcha-button gotcha-flex-end",
        "type": "button",
      });
    refresh.appendChild(
        createElement("img", {
          "class": "gotcha-icon",
          "style": "width: 20px; height: 0.9rem;",
          "src": "data:image/png;base64, iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAYAAABzenr0AAAElklEQVR4nLxXW2xUVRf+1jlnpvD/zFBab1hTSYvGZOgEZg40DVF5ARQvL4gBFEV4wKjRkCjGS8KDt2BMjEYTY4hGCE/VeIlgTCSOvuA0cwZsBQxiIForRit0htrp6XR/Zg1nYNIMZaxTVzIPs/c6+/vWXpe9loMaZM2aNfaJEydWkLwRgAsgDiAcbI+RPCoiRwF8PjQ09Nnx48dHazlXRSbb7OjomNPQ0PAYyc0ArgmWRwD0ksyXDhCJALgBwOxg/4yI7LFt+7l0Ov3blAm4rruW5KsArgLQT3IngA+j0eiRVCpVnKi/ePHiNpJ3kbwnuKGciLyQyWReAWBqJrBs2TInn88r2P0ABkk+0d7evqu7u3v8UtZUkF9N8mUAbQD2Oo6zPp1O5y5JoKura6bv+90AbtMPAWz0PO+PWoEnGDIjl8u9LSIbAHwXDodvPnDgwJ+TEnBddzfJewHsikQim6td9T8V13WfIfk8gC8BrLRtO1osFrcBeEjjWyoUt5B8Sy33PO/Oyfw2BRJvkHwYQJpkTERm4VwAryulYTwev4KkBstJ3/c31BM8Ho//H8AAAL3NThFhxfZgiUAoFNIrmSUiD/T19Z2uB3AQAw9alqUuuAxAGfj8rZMclKVLl0YKhcIpAMc8z0tUKP4rSSaTewCsn0zHGDPPGh0dXQHgf0Ge1wVcJRKJaBq/pGeKSNUU9n1/0DLGrArY7K0XuIpmkOd5TxtjlpM8TXJiXPmHDx8+a1mWpVWr/9ChQyfrSaAsBw8e3O84ziLNgMp1ESnFmkVyLoBfpgO8LOl0uj8ajd4UVMeym0sFzgLQDODX6SSAwCXZbPZJklpjzgD4Wdcdkhstyzo23QTKks1mP00mk9eFQqFSTJRysqurq8n3fWuqdX8q4rruImPM9eoCjI2Nac7+kEgkbv+vCJDcLiLvWsGfVgCNIvJJIpHYoU/ydBMQkRYAv1vB/8svrMu2XC73zcKFC+dNJwGSVwM4VSbQOIHdItu2s67rrpoO8GQy2Q5ACXxrxWIxfRpDE3SU2GySGrEv1tslJO/AOUP3WuFwuPkielaQJU/l8/n36ohvicgmAMMjIyP7tRRXI8CAofaEW7U7qhe667p3a8NN8k19CxwRaSZZCaxWj4vIdt/3X+vt7R2uF/iSJUuax8fHd2jH7DiOluXzpbgsfwHo0QppjJlbT3DFKhaLuwG0isjWnp6ewdKiMeYMybP6UNi2fW2hUNDpJyUij2hDWQ/kWCwWTiaTu0TkVi0+mUzmnfJe1cEkKM1fAVgQdMhbUqlUYSrgnZ2dVwaWLwfwcSQSWVt5ll3to/7+/pHW1tY9xpgObZ1931/X0tIyMDAwcKRWYE3dpqamzSQ/0ksAsLOtre2+ffv2jVXqTTobqotc132cpLoiCqAPgL4b73ue92M10OHh4QXGmNUANgXFZkBEHs1kMh9UA7gUgZJo2x4KhZ4FoEPLnGA5T/J7AEO4MKRqdzUz2P9JrQbwuud5Qxc7uyYCZZk/f35DY2PjLSRXBhOx/mYE276WVhHJGGO+bm9v/6KWefLvAAAA///8j/CcpuGXnQAAAABJRU5ErkJggg=="
        })
      )
    btnGroup.appendChild(refresh)
    var audio = createElement('button', {
      "class": "gotcha-button gotcha-flex-end",
      "type": "button",
      "onclick": "p('gotcha-challenge-audio')"
    })
    audio.appendChild(createElement("img", {
      "class": "gotcha-icon",
      "style": "width: 20px; height: 0.9rem;",
      "src": "data:image/png;base64, iVBORw0KGgoAAAANSUhEUgAAACAAAAAeCAYAAABNChwpAAADS0lEQVR4nLSXT2g7RRTHv292s5GfxPyg2kqgihCKUMEku8ZisFRE6kHxVOzRohYvHvTQgwepKIgoWME/F4v1H1Q8iV5EEEW0Zt1di1AKQmIPak6FtLSmJOs8mWS2hNI2a9J9l3mZeTPvM5P33swauAKpVCqZ8fHxt3O53Ke5XO640WhU486lUZ0Xi8VZIcQGgDt01wmAMd/3/4kz3xzWcT6fT2ez2VcAPA9A9A3dYBjGBIA/EgNwHOcuZv4YQOG88Xa7zXHX+l8ACwsLRr1eX2HmVQDWRXZEdPUApVIpX6vVPiSi+wbZCiHOBbBtW8XKw4ZhzLqu+3sXwHGcz5l5EMgkEd0dw+48AFEoFG7b3t7eA/AAgIkwDD8AcD8ASbZtq4Hb4ywcV0zTnKxWq3+it+uXALxIRM8ycwjgPdXPzEtBEGyIq3QcSbvdlmf7mHmNmVVm/IhenDzTbZM4AQA5AHeqnadSqaUwDN9n5gcB7AghXpZSbiojKeV0IiegYoCIbgIw1+l0vpZSqt22AEwzsyp+h9puLhEAJZ7nfQHgBQBTAJ4AoDJA/RWPEtEvWp9J7ARU6/v+awB+I6KniOhLPfwQM3cBiCiZv6DVakVpKJn5E5V6nU7nRPfdAqCp9ZsTATAM47QOCCF2VJtKpcYARBfUv4kCnClEl5blRAAsyzqtA8w8pdt9ANe0HvltJgJwdHTUv+tFAAc6/ZQcEtGNWj9INAZKpdIygBkAHwGYRy/yvyeiojbdSQygWCw6RPQOgL8Nw3iTiBSMqn5fMfM92tRPBMA0TVUJj5nZJaJ5KeWrAK4DaBDRXwBuRe80fkgMIAiC3SAIKlLKAjM/jl7wrQB4WpvVPM/7KRGAZrPZXwdCrb4lhFCZ8Jj+va6YTCL6mZl/HbSoeowQkRMd32XSX4g8z9ssl8tV13X3HMf5RnfXLctaU4rped5i3J2pN2GtVnudiJ6LO0eJ67rdF7KU8lsiKkspl7e2ttTtONx3geM4TzLzuxc9TC3LuhY5GCRDxYDneesqugHsDzN/ZAAN8R2AewHsnh1Lp9Oxn+Ujf5rZtp0F8FlU6VSp9X3/+qBLKJKR09D3/YNMJvMIgDfUNUBEq3GdK/kvAAD//98qPQMkHPuEAAAAAElFTkSuQmCC"
    }));
    btnGroup.appendChild(audio);
    div.appendChild(btnGroup);
  
    document.getElementById(id).appendChild(div);
    div = null;
    btnGroup = null;
    valButton = undefined;
    refresh = null;
    audio = null;
  }
}

function createElement(type, props) {
  var el = document.createElement(type);
  for (var p in props) {
    el.setAttribute(p, props[p]);
  }
  return el;
}

function p(id) {
  document.getElementById(id).play();
}