import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { Level2Page } from './level2.page';

describe('Level2Page', () => {
  let component: Level2Page;
  let fixture: ComponentFixture<Level2Page>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ Level2Page ],
      schemas: [CUSTOM_ELEMENTS_SCHEMA],
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(Level2Page);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
